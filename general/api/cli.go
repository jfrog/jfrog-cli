package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/urfave/cli"
)

type commandContext interface {
	Args() cli.Args
	String(name string) string
	Bool(name string) bool
	Int(name string) int
	IsSet(name string) bool
	StringSlice(name string) []string
}

const (
	flagVerbose = "verbose"
	flagMethod  = "method"
	flagInput   = "input"
	flagData    = "data"
	flagHeader  = "header"
	flagTimeout = "timeout"
)

// Command runs an authenticated HTTP request against the configured JFrog Platform base URL.
func Command(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	serverDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, true, commonCliUtils.Platform)
	if err != nil {
		return err
	}

	return runApiCmd(c, serverDetails, os.Stdout, os.Stderr)
}

func runApiCmd(c commandContext, serverDetails *coreconfig.ServerDetails, stdOut, stdErr io.Writer) error {
	if serverDetails.GetUrl() == "" {
		return errorutils.CheckErrorf("no JFrog Platform URL specified, either via the --url flag or as part of the server configuration")
	}

	pathArg := c.Args().First()
	fullURL, err := joinPlatformAPIURL(serverDetails.GetUrl(), pathArg)
	if err != nil {
		return err
	}

	method := httpMethodOrDefault(c)
	body, err := resolveRequestBody(c)
	if err != nil {
		return err
	}

	details, err := buildRequestDetails(serverDetails, c)
	if err != nil {
		return err
	}

	timeout := time.Duration(c.Int(flagTimeout)) * time.Second
	client, err := newPlatformHttpClient(serverDetails, timeout)
	if err != nil {
		return err
	}

	return exchangeAndPrint(client, c, method, fullURL, body, details, stdOut, stdErr)
}

func httpMethodOrDefault(c commandContext) string {
	method := strings.ToUpper(strings.TrimSpace(c.String(flagMethod)))
	if method == "" {
		return http.MethodGet
	}
	return method
}

func buildRequestDetails(serverDetails *coreconfig.ServerDetails, c commandContext) (*httputils.HttpClientDetails, error) {
	authDetails, err := serverDetails.CreateAccessAuthConfig()
	if err != nil {
		return nil, err
	}
	httpDetails := authDetails.CreateHttpClientDetails()
	details := &httpDetails
	if err = applyUserHeaders(c, details); err != nil {
		return nil, err
	}
	return details, nil
}

func newPlatformHttpClient(serverDetails *coreconfig.ServerDetails, timeout time.Duration) (*httpclient.HttpClient, error) {
	builder := httpclient.ClientBuilder().
		SetInsecureTls(serverDetails.InsecureTls).
		SetClientCertPath(serverDetails.ClientCertPath).
		SetClientCertKeyPath(serverDetails.ClientCertKeyPath)
	if timeout > 0 {
		builder = builder.SetOverallRequestTimeout(timeout)
	}
	return builder.Build()
}

func exchangeAndPrint(client *httpclient.HttpClient, c commandContext, method, fullURL string, body []byte, details *httputils.HttpClientDetails, stdOut, stdErr io.Writer) error {
	if c.Bool(flagVerbose) {
		writeVerboseRequest(stdErr, method, fullURL, details)
	}

	resp, respBody, _, err := client.Send(method, fullURL, body, true, true, *details, "")
	if err != nil {
		return err
	}
	if resp == nil {
		return errorutils.CheckErrorf("empty response from server")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if c.Bool(flagVerbose) {
		writeVerboseResponse(stdErr, resp, respBody)
	}

	if _, err = fmt.Fprintf(stdErr, "%d\n", resp.StatusCode); err != nil {
		return errorutils.CheckError(err)
	}

	if _, err = stdOut.Write(respBody); err != nil {
		return errorutils.CheckError(err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return cli.NewExitError(fmt.Sprintf("HTTP %d", resp.StatusCode), 1)
	}
	return nil
}

func joinPlatformAPIURL(platformBase, path string) (string, error) {
	base := strings.TrimSuffix(strings.TrimSpace(platformBase), "/")
	p := strings.TrimSpace(path)
	if p == "" {
		return "", errorutils.CheckErrorf("API path must not be empty")
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return base + p, nil
}

func resolveRequestBody(c commandContext) ([]byte, error) {
	inputSet := c.IsSet(flagInput)
	dataSet := c.IsSet(flagData)
	if inputSet && dataSet {
		return nil, errorutils.CheckErrorf("only one of --input and --data can be used")
	}
	if inputSet {
		return readInputPayload(c.String(flagInput))
	}
	if dataSet {
		return []byte(c.String(flagData)), nil
	}
	return nil, nil
}

func readInputPayload(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func applyUserHeaders(c commandContext, details *httputils.HttpClientDetails) error {
	for _, raw := range c.StringSlice(flagHeader) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		key, val, err := parseHeaderKV(raw)
		if err != nil {
			return err
		}
		details.AddHeader(key, val)
	}
	return nil
}

func parseHeaderKV(s string) (key, val string, err error) {
	idx := strings.Index(s, ":")
	if idx <= 0 {
		return "", "", errorutils.CheckErrorf("header %q must use key:value format", s)
	}
	key = strings.TrimSpace(s[:idx])
	val = strings.TrimSpace(s[idx+1:])
	if key == "" {
		return "", "", errorutils.CheckErrorf("header %q must use key:value format", s)
	}
	return key, val, nil
}

func writeVerboseRequest(w io.Writer, method, url string, details *httputils.HttpClientDetails) {
	_, _ = fmt.Fprintf(w, "* Request to %s\n", url)
	_, _ = fmt.Fprintf(w, "> %s\n", method)
	redacted := redactHeaders(details.Headers)
	for k, v := range redacted {
		_, _ = fmt.Fprintf(w, "> %s: %s\n", k, v)
	}
	if !hasHeaderFold(details.Headers, "Authorization") {
		switch {
		case details.AccessToken != "":
			_, _ = fmt.Fprintf(w, "> Authorization: Bearer ***\n")
		case details.User != "" && details.Password != "":
			_, _ = fmt.Fprintf(w, "> Authorization: Basic ***\n")
		}
	}
}

func writeVerboseResponse(w io.Writer, resp *http.Response, body []byte) {
	_, _ = fmt.Fprintf(w, "* Response %s\n", resp.Status)
	for k, vals := range resp.Header {
		for _, v := range vals {
			_, _ = fmt.Fprintf(w, "< %s: %s\n", k, v)
		}
	}
	if len(body) > 0 {
		_, _ = w.Write(body)
		if !bytes.HasSuffix(body, []byte("\n")) {
			_, _ = fmt.Fprintln(w)
		}
	}
}

func hasHeaderFold(h map[string]string, name string) bool {
	for k := range h {
		if strings.EqualFold(k, name) {
			return true
		}
	}
	return false
}

func redactHeaders(h map[string]string) map[string]string {
	if len(h) == 0 {
		return nil
	}
	out := make(map[string]string, len(h))
	for k, v := range h {
		if strings.EqualFold(k, "Authorization") {
			out[k] = redactedAuthValue(v)
			continue
		}
		out[k] = v
	}
	return out
}

func redactedAuthValue(v string) string {
	if v == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(v), "bearer ") {
		return "Bearer ***"
	}
	return "***"
}
