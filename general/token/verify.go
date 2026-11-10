package token

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	generic "github.com/jfrog/jfrog-cli-core/v2/general/token"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/urfave/cli"
	"golang.org/x/term"
)

const (
	stdinPrompt   = "Reading the access token from stdin. Paste it and press Enter (input is hidden): "
	maxTokenInput = 64 * 1024
	utf8Bom       = "\xef\xbb\xbf"
)

var orderedClaims = []string{"jti", "sub", "aud", "scp", "iss", "iat", "exp"}

func AccessTokenVerifyCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	outputFormat, err := commonCliUtils.GetOutputFormat(c, coreformat.Table)
	if err != nil {
		return err
	}
	if outputFormat != coreformat.Table && outputFormat != coreformat.Json {
		return unsupportedVerifyFormat(outputFormat)
	}
	serverDetails, err := verifyServerDetails(c.String("server-id"))
	if err != nil {
		return err
	}
	var readTerminal func(string) (string, error)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		readTerminal = readHiddenLine
	}
	raw, err := readTokenInput(c.Args().Get(0), c.String(cliutils.TokenFile), os.Stdin, readTerminal)
	if err != nil {
		return err
	}
	tokenUnderTest, err := extractToken(raw)
	if err != nil {
		return err
	}

	verifyCmd := generic.NewAccessTokenVerifyCommand().SetServerDetails(serverDetails).SetToken(tokenUnderTest)
	err = commands.Exec(verifyCmd)
	return errors.Join(err, printVerifyResult(verifyCmd.Result(), outputFormat, os.Stdout))
}

func verifyServerDetails(serverIdFlag string) (*coreConfig.ServerDetails, error) {
	serverId := serverIdFlag
	if serverId == "" {
		serverId = os.Getenv(coreutils.ServerID)
	}
	return commands.GetConfig(serverId, true)
}

func readTokenInput(arg, tokenFile string, stdin io.Reader, readTerminal func(prompt string) (string, error)) (string, error) {
	switch {
	case arg != "" && tokenFile != "":
		return "", errorutils.CheckErrorf("provide the token either as an argument or with --token-file, not both")
	case arg != "":
		return arg, nil
	case tokenFile != "":
		file, err := os.Open(tokenFile)
		if err != nil {
			return "", errorutils.CheckErrorf("failed to read token file '%s': %s", tokenFile, err.Error())
		}
		defer func() { _ = file.Close() }()
		return readLimited(file)
	case readTerminal != nil:
		return readTerminal(stdinPrompt)
	default:
		return readLimited(stdin)
	}
}

func readLimited(r io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxTokenInput+1))
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	if len(data) > maxTokenInput {
		return "", errorutils.CheckErrorf("the token input is larger than 64 KiB")
	}
	return string(data), nil
}

// Raw mode avoids the 1,024-byte line limit that canonical terminal mode has on macOS.
func readHiddenLine(prompt string) (string, error) {
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	defer func() { _ = term.Restore(fd, state) }()
	terminal := term.NewTerminal(struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stderr}, "")
	line, err := terminal.ReadPassword(prompt)
	if errors.Is(err, io.EOF) {
		_, _ = fmt.Fprint(os.Stderr, "\r\n")
		return "", nil
	}
	return line, errorutils.CheckError(err)
}

func extractToken(raw string) (string, error) {
	input := strings.TrimSpace(strings.TrimPrefix(raw, utf8Bom))
	var tokenValue string
	if strings.HasPrefix(input, "{") {
		var atcOutput struct {
			AccessToken string `json:"access_token"`
		}
		if err := json.Unmarshal([]byte(input), &atcOutput); err != nil {
			return "", errorutils.CheckErrorf("the token input looks like JSON but isn't valid JSON: %s", jsonErrorWithoutInput(err))
		}
		if atcOutput.AccessToken == "" {
			return "", errorutils.CheckErrorf("the JSON input has no access_token field")
		}
		tokenValue = strings.TrimSpace(atcOutput.AccessToken)
	} else {
		tokenValue = stripTokenDecorations(input)
	}
	if tokenValue == "" {
		return "", errorutils.CheckErrorf("no token provided: pass it as an argument, with --token-file=<path>, or on stdin")
	}
	if strings.ContainsAny(tokenValue, " \t\r\n") {
		return "", errorutils.CheckErrorf("the token input contains more than one value; expected a single token")
	}
	return tokenValue, nil
}

func stripTokenDecorations(input string) string {
	if len(input) >= 2 && strings.HasPrefix(input, `"`) && strings.HasSuffix(input, `"`) {
		input = input[1 : len(input)-1]
	}
	if len(input) >= 7 && strings.EqualFold(input[:7], "bearer ") {
		input = input[7:]
	}
	return strings.TrimSpace(input)
}

func jsonErrorWithoutInput(err error) string {
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return fmt.Sprintf("syntax error at byte %d", syntaxErr.Offset)
	}
	return "unexpected structure"
}

func unsupportedVerifyFormat(outputFormat coreformat.OutputFormat) error {
	return errorutils.CheckErrorf("unsupported format '%s' for access-token-verify. Accepted values: table, json", outputFormat)
}

func printVerifyResult(result *generic.AccessTokenVerifyResult, outputFormat coreformat.OutputFormat, w io.Writer) error {
	switch outputFormat {
	case coreformat.Json:
		if result == nil {
			return nil
		}
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return errorutils.CheckError(err)
		}
		_, err = fmt.Fprintln(w, string(data))
		return errorutils.CheckError(err)
	case coreformat.Table:
		return printVerifyTable(result, w)
	default:
		return unsupportedVerifyFormat(outputFormat)
	}
}

func printVerifyTable(result *generic.AccessTokenVerifyResult, w io.Writer) error {
	if result == nil || result.Payload == nil {
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "FIELD\tVALUE")
	for _, row := range payloadRows(result.Payload) {
		_, _ = fmt.Fprintf(tw, "%s\t%s\n", row[0], row[1])
	}
	if err := tw.Flush(); err != nil {
		return errorutils.CheckError(err)
	}
	if !result.Valid {
		return nil
	}
	_, _ = fmt.Fprintln(w, "\nTYPE")
	for _, tokenType := range result.Types {
		_, _ = fmt.Fprintln(w, tokenType)
	}
	return nil
}

func payloadRows(payload map[string]interface{}) [][2]string {
	var rows [][2]string
	for _, key := range orderedClaims {
		value, ok := payload[key]
		switch {
		case key == "exp" && !ok:
			rows = append(rows, [2]string{key, "(no expiry)"})
		case !ok:
		case key == "iat" || key == "exp":
			rows = append(rows, [2]string{key, formatEpoch(value)})
		default:
			rows = append(rows, [2]string{key, formatClaim(value)})
		}
	}
	for _, key := range sortedKeys(payload) {
		if key == "ext" || isOrderedClaim(key) {
			continue
		}
		rows = append(rows, [2]string{key, formatClaim(payload[key])})
	}
	if ext, ok := payload["ext"].(map[string]interface{}); ok {
		for _, key := range sortedKeys(ext) {
			rows = append(rows, [2]string{"ext." + key, formatClaim(ext[key])})
		}
	}
	return rows
}

func formatEpoch(value interface{}) string {
	epoch, err := strconv.ParseInt(fmt.Sprint(value), 10, 64)
	if err != nil {
		return formatClaim(value)
	}
	return fmt.Sprintf("%s (epoch: %d)", time.Unix(epoch, 0).UTC().Format("2006-01-02 15:04:05 UTC"), epoch)
}

func formatClaim(value interface{}) string {
	if values, ok := value.([]interface{}); ok {
		parts := make([]string, 0, len(values))
		for _, v := range values {
			parts = append(parts, fmt.Sprint(v))
		}
		return strings.Join(parts, ", ")
	}
	if nested, ok := value.(map[string]interface{}); ok {
		data, err := json.Marshal(nested)
		if err == nil {
			return string(data)
		}
	}
	return fmt.Sprint(value)
}

func isOrderedClaim(key string) bool {
	for _, claim := range orderedClaims {
		if claim == key {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
