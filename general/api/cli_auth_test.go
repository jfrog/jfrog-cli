package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

// fakeJWT returns an unsigned JWT-shaped token that expires after ttl.
func fakeJWT(t *testing.T, ttl time.Duration) string {
	t.Helper()
	enc := func(v any) string {
		b, err := json.Marshal(v)
		require.NoError(t, err)
		return base64.RawURLEncoding.EncodeToString(b)
	}
	now := time.Now().Unix()
	return enc(map[string]string{"alg": "RS256", "typ": "JWT"}) + "." +
		enc(map[string]any{"sub": "jfac@01/users/tester", "iat": now, "exp": now + int64(ttl.Seconds())}) + ".sig"
}

// A server saved by `jf c add --user --password` (refreshable tokens enabled, the default) must send
// the refreshable access token as a bearer token, like every other command, and not basic auth.
// A server saved with --basic-auth-only keeps using basic auth.
func TestCommand_UserPasswordServerAuth(t *testing.T) {
	token := fakeJWT(t, time.Hour)
	tests := []struct {
		name              string
		tokenRefreshInMin int
		wantAuth          string
	}{
		{name: "refreshable tokens (default)", tokenRefreshInMin: 60, wantAuth: "Bearer " + token},
		{name: "basic auth only", tokenRefreshInMin: -1, wantAuth: "Basic " + base64.StdEncoding.EncodeToString([]byte("tester:password"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotAuth string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/artifactory/api/security/token":
					w.Header().Set("Content-Type", "application/json")
					_, _ = fmt.Fprintf(w, `{"access_token":%q,"refresh_token":"refresh","expires_in":3600}`, token)
				case "/test/api":
					gotAuth = r.Header.Get("Authorization")
					w.WriteHeader(http.StatusOK)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			t.Cleanup(srv.Close)

			home := t.TempDir()
			t.Setenv("JFROG_CLI_HOME_DIR", home)
			t.Setenv("JFROG_CLI_REPORT_USAGE", "false")
			t.Setenv("CI", "true")
			conf := fmt.Sprintf(`{"servers":[{"serverId":"basic","url":%q,"artifactoryUrl":%q,"user":"tester","password":"password","tokenRefreshInterval":%d,"isDefault":true}],"version":"6"}`,
				srv.URL+"/", srv.URL+"/artifactory/", tt.tokenRefreshInMin)
			require.NoError(t, os.WriteFile(filepath.Join(home, "jfrog-cli.conf.v6"), []byte(conf), 0o600))

			app := cli.NewApp()
			app.Commands = []cli.Command{{Name: "api", Flags: cliutils.GetCommandFlags(cliutils.Api), Action: Command}}
			require.NoError(t, app.Run([]string{"jf", "api", "/test/api", "--server-id", "basic"}))
			assert.Equal(t, tt.wantAuth, gotAuth)
		})
	}
}
