package token

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	generic "github.com/jfrog/jfrog-cli-core/v2/general/token"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type untouchedReader struct{ t *testing.T }

func (r untouchedReader) Read([]byte) (int, error) {
	r.t.Error("stdin was read")
	return 0, io.EOF
}

func TestExtractToken(t *testing.T) {
	atcOutput := "{\n  \"scope\": \"applied-permissions/user\",\n  \"access_token\": \"a.b.c\",\n  \"expires_in\": 3600,\n  \"token_type\": \"Bearer\"\n}\n"
	for _, raw := range []string{
		"a.b.c", "  a.b.c\n", "Bearer a.b.c", "bearer a.b.c", `"a.b.c"`, "\ufeffa.b.c\r\n",
		atcOutput,
		"\ufeff" + `{"access_token":"a.b.c","reference_token":"cmVmdGtuOjAx"}`,
	} {
		token, err := extractToken(raw)
		require.NoError(t, err, raw)
		assert.Equal(t, "a.b.c", token, raw)
	}

	for _, tc := range []struct {
		raw     string
		wantErr string
	}{
		{`{"token_id":"x"}`, "the JSON input has no access_token field"},
		{`{"access_token":""}`, "the JSON input has no access_token field"},
		{`{not json a.b.c`, "the token input looks like JSON but isn't valid JSON"},
		{"", "no token provided"},
		{"   \n", "no token provided"},
		{"a.b.c\nd.e.f", "the token input contains more than one value; expected a single token"},
	} {
		_, err := extractToken(tc.raw)
		assert.ErrorContains(t, err, tc.wantErr, tc.raw)
		if err != nil {
			assert.NotContains(t, err.Error(), "a.b.c")
		}
	}
}

func TestReadTokenInput(t *testing.T) {
	failTerminal := func(string) (string, error) {
		t.Error("terminal was read")
		return "", nil
	}
	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "token.txt")
	require.NoError(t, os.WriteFile(tokenFile, []byte("x\n"), 0o600))

	raw, err := readTokenInput("a.b.c", "", untouchedReader{t}, failTerminal)
	require.NoError(t, err)
	assert.Equal(t, "a.b.c", raw)

	raw, err = readTokenInput("", tokenFile, untouchedReader{t}, failTerminal)
	require.NoError(t, err)
	assert.Equal(t, "x\n", raw)

	_, err = readTokenInput("a.b.c", tokenFile, untouchedReader{t}, failTerminal)
	assert.EqualError(t, err, "provide the token either as an argument or with --token-file, not both")

	_, err = readTokenInput("", "missing.txt", untouchedReader{t}, failTerminal)
	assert.ErrorContains(t, err, "failed to read token file 'missing.txt'")

	raw, err = readTokenInput("", "", strings.NewReader("a.b.c\n"), nil)
	require.NoError(t, err)
	assert.Equal(t, "a.b.c\n", raw)

	var gotPrompt string
	raw, err = readTokenInput("", "", untouchedReader{t}, func(prompt string) (string, error) {
		gotPrompt = prompt
		return "a.b.c", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "a.b.c", raw)
	assert.Equal(t, stdinPrompt, gotPrompt)

	_, err = readTokenInput("", "", strings.NewReader(strings.Repeat("x", 64*1024+1)), nil)
	assert.EqualError(t, err, "the token input is larger than 64 KiB")
}

func TestVerifyServerDetails(t *testing.T) {
	t.Setenv(coreutils.HomeDir, t.TempDir())
	t.Setenv(coreutils.ServerID, "")
	details, err := verifyServerDetails("")
	require.NoError(t, err)
	assert.Empty(t, details.Url)

	require.NoError(t, coreConfig.SaveServersConf([]*coreConfig.ServerDetails{
		{ServerId: "a", Url: "https://a.example/", IsDefault: true},
		{ServerId: "b", Url: "https://b.example/"},
	}))
	for _, tc := range []struct {
		flag, env, want string
	}{
		{"", "", "a"},
		{"b", "", "b"},
		{"", "b", "b"},
	} {
		t.Setenv(coreutils.ServerID, tc.env)
		details, err = verifyServerDetails(tc.flag)
		require.NoError(t, err)
		assert.Equal(t, tc.want, details.ServerId)
		assert.False(t, details.InsecureTls)
	}
}

func TestAccessTokenVerifyFlags(t *testing.T) {
	var names []string
	for _, flag := range cliutils.GetCommandFlags(cliutils.AccessTokenVerify) {
		names = append(names, flag.GetName())
	}
	assert.ElementsMatch(t, []string{"server-id", "token-file", "format"}, names)
}

func sampleVerifyResult(valid bool) *generic.AccessTokenVerifyResult {
	return &generic.AccessTokenVerifyResult{
		Valid: valid,
		Payload: map[string]interface{}{
			"jti": "id-1",
			"sub": "jfac@01h/users/alice",
			"aud": "*@*",
			"scp": "applied-permissions/user",
			"iss": "jfac@01h",
			"iat": json.Number("1700000000"),
			"exp": json.Number("1700003600"),
			"ext": map[string]interface{}{"revocable": "true"},
		},
		Types: []generic.TokenType{generic.TokenTypeIdentity},
	}
}

func TestPrintVerifyResult_Table_Valid(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, printVerifyResult(sampleVerifyResult(true), coreformat.Table, &buf))
	for _, want := range []string{"FIELD", "sub", "UTC (epoch: 1700000000)", "ext.revocable", "TYPE", "Identity Token"} {
		assert.Contains(t, buf.String(), want)
	}
}

func TestPrintVerifyResult_Table_NoExpiry(t *testing.T) {
	result := sampleVerifyResult(true)
	delete(result.Payload, "exp")
	var buf bytes.Buffer
	require.NoError(t, printVerifyResult(result, coreformat.Table, &buf))
	assert.Contains(t, buf.String(), "(no expiry)")
}

func TestPrintVerifyResult_Table_Invalid(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, printVerifyResult(sampleVerifyResult(false), coreformat.Table, &buf))
	assert.Contains(t, buf.String(), "sub")
	assert.NotContains(t, buf.String(), "TYPE")
}

func TestPrintVerifyResult_JSON(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, printVerifyResult(sampleVerifyResult(true), coreformat.Json, &buf))
	var out struct {
		Valid   bool                   `json:"valid"`
		Types   []string               `json:"types"`
		Payload map[string]interface{} `json:"payload"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.True(t, out.Valid)
	assert.Equal(t, []string{"Identity Token"}, out.Types)
	assert.IsType(t, map[string]interface{}{}, out.Payload["ext"])
}

func TestPrintVerifyResult_JSON_ParseFailure(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, printVerifyResult(&generic.AccessTokenVerifyResult{Error: "failed to parse token: x"}, coreformat.Json, &buf))
	assert.JSONEq(t, `{"valid":false,"error":"failed to parse token: x"}`, buf.String())
}

func TestPrintVerifyResult_UnsupportedFormat(t *testing.T) {
	err := printVerifyResult(sampleVerifyResult(true), "xml", &bytes.Buffer{})
	assert.EqualError(t, err, "unsupported format 'xml' for access-token-verify. Accepted values: table, json")
}
