package api

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/agnivade/levenshtein"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
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

// Per-field contains-match weights. They sum (an operation matching in two
// fields ranks above one matching in only its strongest field) and are capped
// at 100. Fuzzy-fallback scores are confined to fuzzyScoreMax and below, so an
// exact/substring hit on even the weakest field (tag) always outranks a pure
// fuzzy hit.
const (
	weightOperationId  = 40
	weightPath         = 30
	weightSummary      = 20
	weightTag          = 10
	fuzzyScoreMax      = 9
	fuzzySimilarityMin = 0.4
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

	defaultFormat := coreformat.Table
	if corecommon.AIHelpEnabled() {
		defaultFormat = coreformat.Json
	}
	outputFormat, err := commonCliUtils.GetOutputFormat(c, defaultFormat)
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
			Method:  op.Method,
			Path:    op.Path,
			Summary: op.Summary,
			Tags:    op.Tags,
			Score:   score,
			JfApi:   jfApiOneLiner(op),
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
	if best < fuzzySimilarityMin {
		return 0
	}
	return min(max(int(best*fuzzyScoreMax), 1), fuzzyScoreMax)
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

// jfApiOneliner is a ready-to-run `jf api` invocation for op: the method is
// omitted for GET, since that's runApiCmd's own default.
func jfApiOneLiner(op apispec.Operation) string {
	if op.Method == "GET" {
		return fmt.Sprintf("jf api %s", op.Path)
	}
	return fmt.Sprintf("jf api %s -X %s", op.Path, op.Method)
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
	_, _ = fmt.Fprintln(tw, "METHOD\tPATH\tSUMMARY\tTAGS\tSCORE\tJF API")
	for _, m := range result.Matches {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\n", m.Method, m.Path, m.Summary, strings.Join(m.Tags, ","), m.Score, m.JfApi)
	}
	return tw.Flush()
}
