package api

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/agnivade/levenshtein"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	apispec "github.com/jfrog/jfrog-cli/docs/api-spec"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

const (
	flagTag   = "tag"
	flagLimit = "limit"
)

// envFuzzySimilarityMin overrides fuzzySimilarityMin for experimentation --
// accepts any value parseable as a float64 in [0, 1]. Unset or invalid falls
// back to the default, same convention as corecommon.AIHelpEnabled's handling
// of JFROG_CLI_AI_HELP.
const envFuzzySimilarityMin = "JFROG_CLI_API_DOCS_SEARCH_FUZZY_MIN"

// Per-field contains-match weights. They sum (an operation matching in two
// fields ranks above one matching in only its strongest field) and are capped
// at 100. Fuzzy-fallback scores are confined to fuzzyScoreMax and below, so an
// exact/substring hit on even the weakest field (tag) always outranks a pure
// fuzzy hit.
const (
	weightOperationId = 40
	weightPath        = 30
	weightSummary     = 20
	weightTag         = 10
	fuzzyScoreMax     = 9
	// fuzzySimilarityMin is deliberately high: whole-string Levenshtein
	// similarity between two semantically unrelated but coincidentally
	// similar-length words (e.g. "evidence" vs "environments", 0.42) crosses a
	// low bar easily, flooding results with false positives that also mask a
	// legitimate empty-result response. 0.6 still catches real typos
	// ("usres"->"users" 0.6, "workrs"->"workers" 0.86) while rejecting
	// unrelated words of similar length ("evidence"->"environments" 0.42,
	// "token"->"worker" 0.5, "scan"->"scim" 0.5).
	fuzzySimilarityMin = 0.6
	defaultLimit       = 10
)

// match is a single scored, ranked apispec.Operation ready for rendering.
type match struct {
	Method  string   `json:"method"`
	Path    string   `json:"path"`
	Summary string   `json:"summary"`
	Tags    []string `json:"tags"`
	Score   int      `json:"score"`
	JfApi   string   `json:"jf_api"`
	// Parameters and RequestBody are the payload/parameter data an agent needs
	// to actually call this operation, not just find it -- omitted when absent
	// (e.g. Parameters for a path with none, RequestBody for GET/DELETE).
	Parameters  []apispec.Parameter  `json:"parameters,omitempty"`
	RequestBody *apispec.RequestBody `json:"request_body,omitempty"`
}

// searchResult is the JSON/table rendering payload for `jf api docs search`.
type searchResult struct {
	SpecBundle  string  `json:"spec_bundle"`
	SpecVersion string  `json:"spec_version"`
	Query       string  `json:"query"`
	Matches     []match `json:"matches"`
	Message     string  `json:"message,omitempty"`
}

// SearchCommand implements `jf api docs search <query>`. It ranks operations
// from the embedded OpenAPI spec bundle by relevance to query -- this is a
// local, offline lookup; no server configuration or network call is involved.
func SearchCommand(c *cli.Context) error {
	return runSearchCmd(c, os.Stdout)
}

// runSearchCmd is split out from SearchCommand so tests can supply their own
// stdOut without hijacking the real os.Stdout -- same split as
// Command/runApiCmd in cli.go. JSON rendering still goes through the shared
// client logger's Output channel (see renderJSON), matching
// printTokenResponse's convention in general/token/cli.go.
func runSearchCmd(c *cli.Context, stdOut io.Writer) error {
	if c.NArg() != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	query := c.Args().First()
	tag := c.String(flagTag)
	method := c.String(flagMethod)
	limit := c.Int(flagLimit)
	if limit <= 0 {
		limit = defaultLimit
	}

	ops, err := apispec.Operations()
	if err != nil {
		return errorutils.CheckError(err)
	}

	matches := filterAndScore(ops, query, tag, method)
	if len(matches) > limit {
		matches = matches[:limit]
	}

	info := apispec.Info()
	result := searchResult{
		SpecBundle:  info.SpecBundle,
		SpecVersion: info.SpecVersion,
		Query:       query,
		Matches:     matches,
	}
	if len(matches) == 0 {
		result.Message = fmt.Sprintf(
			"No matching operations found in the embedded %q OpenAPI spec bundle for query %q. "+
				"The bundle may be incomplete (see spec_bundle) -- try 'jf api <path>' directly if you already know the endpoint.",
			result.SpecBundle, query)
	}

	// JSON is the default output format -- this command exists primarily for
	// agent consumption; --format table is available for humans who want it.
	outputFormat, err := commonCliUtils.GetOutputFormat(c, coreformat.Json)
	if err != nil {
		return err
	}

	switch outputFormat {
	case coreformat.Json:
		return renderJSON(result)
	case coreformat.Table:
		return renderTable(result, stdOut)
	default:
		return errorutils.CheckErrorf("unsupported format '%s' for api docs search. Accepted values: table, json", outputFormat)
	}
}

// filterAndScore applies the --tag/--method hard filters, scores every
// remaining operation against query, drops zero-scoring operations, and
// returns matches sorted best-first (score desc, then path/method asc).
func filterAndScore(ops []apispec.Operation, query, tag, method string) []match {
	q := strings.ToLower(strings.TrimSpace(query))
	tagFilter := strings.ToLower(strings.TrimSpace(tag))
	methodFilter := strings.ToUpper(strings.TrimSpace(method))

	var matches []match
	for _, op := range ops {
		if methodFilter != "" && op.Method != methodFilter {
			continue
		}
		if tagFilter != "" && !hasTag(op.Tags, tagFilter) {
			continue
		}
		score := scoreOperation(op, q)
		if score == 0 {
			continue
		}
		matches = append(matches, match{
			Method:      op.Method,
			Path:        op.Path,
			Summary:     op.Summary,
			Tags:        op.Tags,
			Score:       score,
			JfApi:       jfApiOneLiner(op),
			Parameters:  op.Parameters,
			RequestBody: op.RequestBody,
		})
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		if matches[i].Path != matches[j].Path {
			return matches[i].Path < matches[j].Path
		}
		return matches[i].Method < matches[j].Method
	})
	return matches
}

// scoreOperation returns 0-100. An empty query matches every operation with
// the maximum score (a convenience "list everything" behavior, not a
// documented feature). A non-empty query that neither contains-matches nor
// clears the fuzzy similarity threshold on any field scores 0 (excluded).
func scoreOperation(op apispec.Operation, q string) int {
	if q == "" {
		return weightOperationId + weightPath + weightSummary + weightTag
	}

	score := 0
	matchedAnyField := false
	if strings.Contains(strings.ToLower(op.OperationId), q) {
		score += weightOperationId
		matchedAnyField = true
	}
	if strings.Contains(strings.ToLower(op.Path), q) {
		score += weightPath
		matchedAnyField = true
	}
	if strings.Contains(strings.ToLower(op.Summary), q) {
		score += weightSummary
		matchedAnyField = true
	}
	for _, t := range op.Tags {
		if strings.Contains(strings.ToLower(t), q) {
			score += weightTag
			matchedAnyField = true
			break
		}
	}
	if matchedAnyField {
		return min(score, 100)
	}

	return fuzzyScore(op, q)
}

// fuzzyScore is only consulted when no field contains q as a substring. It
// returns a value in [0, fuzzyScoreMax], where 0 means "too dissimilar to
// count as a match at all" (excludes the operation from results).
func fuzzyScore(op apispec.Operation, q string) int {
	fields := append([]string{op.OperationId, op.Path, op.Summary}, op.Tags...)
	best := 0.0
	for _, field := range fields {
		if sim := similarity(q, strings.ToLower(field)); sim > best {
			best = sim
		}
	}
	if best < fuzzySimilarityThreshold() {
		return 0
	}
	return min(max(int(best*fuzzyScoreMax), 1), fuzzyScoreMax)
}

// fuzzySimilarityThreshold resolves the effective fuzzy-match floor: the
// JFROG_CLI_API_DOCS_SEARCH_FUZZY_MIN env var when it parses as a float64 in
// [0, 1], otherwise fuzzySimilarityMin.
func fuzzySimilarityThreshold() float64 {
	v := strings.TrimSpace(os.Getenv(envFuzzySimilarityMin))
	if v == "" {
		return fuzzySimilarityMin
	}
	parsed, err := strconv.ParseFloat(v, 64)
	if err != nil || parsed < 0 || parsed > 1 {
		return fuzzySimilarityMin
	}
	return parsed
}

// similarity is a normalized Levenshtein similarity in [0, 1]; 1 means
// identical strings, 0 means maximally different.
func similarity(a, b string) float64 {
	maxLen := max(len(a), len(b))
	if maxLen == 0 {
		return 1
	}
	dist := levenshtein.ComputeDistance(a, b)
	return 1 - float64(dist)/float64(maxLen)
}

func hasTag(tags []string, tagFilter string) bool {
	for _, t := range tags {
		if strings.Contains(strings.ToLower(t), tagFilter) {
			return true
		}
	}
	return false
}

// jfApiOneLiner is a ready-to-run `jf api` invocation for op: the method is
// omitted for GET, since that's runApiCmd's own default. When op has a
// request body with required fields, a minimal placeholder JSON payload is
// appended via -d so the command is runnable (after filling in real values),
// not just a bare skeleton the caller has to guess the shape of.
func jfApiOneLiner(op apispec.Operation) string {
	var b strings.Builder
	b.WriteString("jf api ")
	b.WriteString(op.Path)
	if op.Method != "GET" {
		b.WriteString(" -X ")
		b.WriteString(op.Method)
	}
	if skeleton := requestBodySkeleton(op.RequestBody); skeleton != "" {
		b.WriteString(` -H "Content-Type: application/json" -d '`)
		b.WriteString(skeleton)
		b.WriteString(`'`)
	}
	return b.String()
}

// requestBodySkeleton renders a minimal JSON object covering just rb's
// required fields with type-appropriate placeholder values, or "" when rb is
// nil or has no required fields (an all-optional body isn't worth guessing at
// in a one-liner -- the full field list is still in the match's request_body).
func requestBodySkeleton(rb *apispec.RequestBody) string {
	if rb == nil {
		return ""
	}
	var fields []string
	for _, p := range rb.Properties {
		if p.Required {
			fields = append(fields, fmt.Sprintf("%q:%s", p.Name, placeholderJSONValue(p.Type)))
		}
	}
	if len(fields) == 0 {
		return ""
	}
	return "{" + strings.Join(fields, ",") + "}"
}

// placeholderJSONValue returns a type-appropriate placeholder JSON literal.
// A referenced schema name (e.g. "BuildTarget") isn't a JSON-schema
// primitive, so it falls back to the empty-string placeholder like any other
// unrecognized type.
func placeholderJSONValue(propType string) string {
	switch {
	case propType == "boolean":
		return "false"
	case propType == "integer" || propType == "number":
		return "0"
	case propType == "array" || strings.HasPrefix(propType, "array<"):
		return "[]"
	case propType == "object":
		return "{}"
	default:
		return `""`
	}
}

// renderJSON writes result as indented JSON via the shared client logger --
// same pattern as printTokenResponse's JSON branch in general/token/cli.go.
func renderJSON(result searchResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return errorutils.CheckErrorf("failed to marshal api docs search result: %s", err.Error())
	}
	log.Output(clientUtils.IndentJson(data))
	return nil
}

// renderTable writes result as a tabwriter-rendered table to w, mirroring
// printTokenTable's construction style in general/token/cli.go.
func renderTable(result searchResult, w io.Writer) error {
	if len(result.Matches) == 0 {
		_, err := fmt.Fprintf(w, "%s (spec_bundle=%s)\n", result.Message, result.SpecBundle)
		return errorutils.CheckError(err)
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "METHOD\tPATH\tSUMMARY\tTAGS\tSCORE\tPARAMS\tBODY\tJF API")
	for _, m := range result.Matches {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\n",
			m.Method, m.Path, m.Summary, strings.Join(m.Tags, ","), m.Score,
			formatParams(m.Parameters), formatRequestBody(m.RequestBody), m.JfApi)
	}
	return tw.Flush()
}

// formatParams and formatRequestBody render a compact, single-line summary of
// field names for the table view (a "*" suffix marks a required field) --
// the JSON view carries the full type/description/default detail instead.
func formatParams(params []apispec.Parameter) string {
	if len(params) == 0 {
		return "-"
	}
	names := make([]string, len(params))
	for i, p := range params {
		names[i] = requiredMark(p.Name, p.Required)
	}
	return strings.Join(names, ",")
}

func formatRequestBody(rb *apispec.RequestBody) string {
	if rb == nil || len(rb.Properties) == 0 {
		return "-"
	}
	names := make([]string, len(rb.Properties))
	for i, p := range rb.Properties {
		names[i] = requiredMark(p.Name, p.Required)
	}
	return strings.Join(names, ",")
}

func requiredMark(name string, required bool) string {
	if required {
		return name + "*"
	}
	return name
}
