package api

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	apispec "github.com/jfrog/jfrog-cli/docs/api-spec"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

// describeResult is the JSON/table rendering payload for `jf api docs describe`.
type describeResult struct {
	SpecBundle  string               `json:"spec_bundle"`
	SpecVersion string               `json:"spec_version"`
	Method      string               `json:"method"`
	Path        string               `json:"path"`
	Summary     string               `json:"summary,omitempty"`
	Tags        []string             `json:"tags,omitempty"`
	Parameters  []apispec.Parameter  `json:"parameters,omitempty"`
	RequestBody *apispec.RequestBody `json:"request_body,omitempty"`
	Responses   []apispec.Response   `json:"responses,omitempty"`
	JfApi       string               `json:"jf_api"`
}

// DescribeCommand implements `jf api docs describe <method> <path>`. It returns
// the full trimmed operation view (parameters, request body schema, response
// codes/descriptions, example payload when declared, and a ready-to-run jf api
// one-liner) for a single method+path pulled from the embedded OpenAPI spec
// bundle -- same local, offline lookup model as `jf api docs search`.
func DescribeCommand(c *cli.Context) error {
	return runDescribeCmd(c, os.Stdout)
}

// runDescribeCmd is split out from DescribeCommand so tests can supply their
// own stdOut without hijacking the real os.Stdout -- same split as
// Command/runApiCmd in cli.go and SearchCommand/runSearchCmd in docs_search.go.
func runDescribeCmd(c *cli.Context, stdOut io.Writer) error {
	if c.NArg() != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	method := c.Args().Get(0)
	path := normalizeApiPath(c.Args().Get(1))
	if err := maybeRequireFullApiDocsBundle(); err != nil {
		return err
	}

	info := apispec.Info()
	op, ok := apispec.FindOperation(method, path)
	if !ok {
		return errorutils.CheckErrorf(
			"no operation found for %s %s in the embedded %q OpenAPI spec bundle. "+
				"Run 'jf api docs search <query>' to find the exact method/path first -- "+
				"the bundle may be incomplete (see spec_bundle), or the path may need to "+
				"match the catalog's literal {param} placeholders exactly.",
			strings.ToUpper(strings.TrimSpace(method)), path, info.SpecBundle)
	}

	result := describeResult{
		SpecBundle:  info.SpecBundle,
		SpecVersion: info.SpecVersion,
		Method:      op.Method,
		Path:        op.Path,
		Summary:     op.Summary,
		Tags:        op.Tags,
		Parameters:  op.Parameters,
		RequestBody: op.RequestBody,
		Responses:   op.Responses,
		JfApi:       jfApiOneLiner(op),
	}

	// JSON is the unconditional default -- this command exists primarily for
	// agent consumption, matching `jf api docs search`'s convention.
	outputFormat, err := commonCliUtils.GetOutputFormat(c, coreformat.Json)
	if err != nil {
		return err
	}

	switch outputFormat {
	case coreformat.Json:
		return renderDescribeJSON(result)
	case coreformat.Table:
		return renderDescribeTable(result, stdOut)
	default:
		return errorutils.CheckErrorf("unsupported format '%s' for api docs describe. Accepted values: table, json", outputFormat)
	}
}

// normalizeApiPath prepends a leading "/" when missing, same convention as
// joinPlatformAPIURL in cli.go, so both "GET access/api/v2/users" and
// "GET /access/api/v2/users" resolve against the catalog.
func normalizeApiPath(path string) string {
	p := strings.TrimSpace(path)
	if p != "" && !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

// renderDescribeJSON writes result as indented JSON via the shared client
// logger -- same pattern as renderJSON in docs_search.go.
func renderDescribeJSON(result describeResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return errorutils.CheckErrorf("failed to marshal api docs describe result: %s", err.Error())
	}
	log.Output(clientUtils.IndentJson(data))
	return nil
}

// renderDescribeTable writes result as a single-record key/value dump to w --
// there is exactly one operation to show, not a ranked list, so this doesn't
// reuse renderTable's multi-row shape from docs_search.go.
func renderDescribeTable(result describeResult, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "METHOD\t%s\n", result.Method)
	_, _ = fmt.Fprintf(tw, "PATH\t%s\n", result.Path)
	_, _ = fmt.Fprintf(tw, "SUMMARY\t%s\n", result.Summary)
	_, _ = fmt.Fprintf(tw, "TAGS\t%s\n", strings.Join(result.Tags, ","))
	_, _ = fmt.Fprintf(tw, "PARAMETERS\t%s\n", formatParams(result.Parameters))
	_, _ = fmt.Fprintf(tw, "REQUEST BODY\t%s\n", formatRequestBody(result.RequestBody))
	if result.RequestBody != nil && len(result.RequestBody.Example) > 0 {
		_, _ = fmt.Fprintf(tw, "REQUEST BODY EXAMPLE\t%s\n", string(result.RequestBody.Example))
	}
	_, _ = fmt.Fprintf(tw, "RESPONSES\t%s\n", formatResponses(result.Responses))
	_, _ = fmt.Fprintf(tw, "JF API\t%s\n", result.JfApi)
	return tw.Flush()
}

// formatResponses renders a compact "code:description, ..." summary of an
// operation's declared responses for the table view.
func formatResponses(responses []apispec.Response) string {
	if len(responses) == 0 {
		return "-"
	}
	parts := make([]string, len(responses))
	for i, r := range responses {
		if r.Description == "" {
			parts[i] = r.Code
			continue
		}
		parts[i] = r.Code + ":" + r.Description
	}
	return strings.Join(parts, ", ")
}
