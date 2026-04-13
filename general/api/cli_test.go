package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	testhelpers "github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestJoinPlatformAPIURL(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		path    string
		want    string
		wantErr error
	}{
		{
			name: "base with trailing slash and path with leading slash",
			base: "https://acme.jfrog.io/",
			path: "/access/api/v1/users",
			want: "https://acme.jfrog.io/access/api/v1/users",
		},
		{
			name: "base without trailing slash and path without leading slash",
			base: "https://acme.jfrog.io",
			path: "access/api/v1/users",
			want: "https://acme.jfrog.io/access/api/v1/users",
		},
		{
			name:    "empty path returns error",
			base:    "https://x.io",
			path:    "",
			wantErr: errorutils.CheckErrorf("API path must not be empty"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := joinPlatformAPIURL(tt.base, tt.path)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, u)
			}
		})
	}
}

func TestParseHeaderKV(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantK   string
		wantV   string
		wantErr error
	}{
		{
			name:  "valid header",
			in:    "Content-Type: application/json",
			wantK: "Content-Type",
			wantV: "application/json",
		},
		{
			name:    "missing colon",
			in:      "bad",
			wantErr: errorutils.CheckErrorf(`header "bad" must use key:value format`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, v, err := parseHeaderKV(tt.in)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantK, k)
			assert.Equal(t, tt.wantV, v)
		})
	}
}

func TestHasHeaderFold(t *testing.T) {
	tests := []struct {
		name string
		hdr  map[string]string
		key  string
		want bool
	}{
		{
			name: "match case insensitive",
			hdr:  map[string]string{"Authorization": "x"},
			key:  "authorization",
			want: true,
		},
		{
			name: "no match",
			hdr:  map[string]string{"X-Other": "y"},
			key:  "authorization",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasHeaderFold(tt.hdr, tt.key))
		})
	}
}

func TestResolveRequestBody(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		input   string
		want    []byte
		wantErr error
	}{
		{
			name: "data flag",
			args: []string{"cmd", "-d", `{"x":1}`},
			want: []byte(`{"x":1}`),
		},
		{
			name:  "input flag",
			args:  []string{"cmd"},
			input: "input content",
			want:  []byte("input content"),
		},
		{
			name:    "input and data flags mutually exclusive",
			args:    []string{"cmd", "--input", "/dev/null", "-d", `{"x":1}`},
			wantErr: errorutils.CheckErrorf("only one of --input and --data can be used"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := cli.NewApp()
			app.Flags = []cli.Flag{
				cli.StringFlag{Name: "input"},
				cli.StringFlag{Name: "data, d"},
			}
			var b []byte
			var err error
			app.Action = func(c *cli.Context) error {
				b, err = resolveRequestBody(c)
				return nil
			}
			if tt.input != "" {
				tempFile := testhelpers.CreateTempFile(t, tt.input)
				tt.args = append(tt.args, "--input", tempFile)
			}

			require.NoError(t, app.Run(tt.args))

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, b)
			}
		})
	}
}

func TestApi(t *testing.T) {
	type mockConfig struct {
		handler  func(http.ResponseWriter, *http.Request)
		path     string
		response []byte
		status   int
	}
	tests := []struct {
		name         string
		args         commandArgs
		server       mockConfig
		wantResponse []byte
		wantStatus   int
		wantErr      error
	}{
		{
			name: "success",
			args: commandArgs{
				path: "/success",
			},
			server: mockConfig{
				path:     "/success",
				status:   200,
				response: []byte("OK"),
			},
			wantStatus:   200,
			wantResponse: []byte("OK"),
		},
		{
			name: "default method is GET",
			args: commandArgs{
				path: "/default-method",
			},
			server: mockConfig{
				path: "/default-method",
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write([]byte(r.Method)); err != nil { // #nosec
						t.Log(err)
					}
				},
			},
			wantStatus:   200,
			wantResponse: []byte("GET"),
		},
		{
			name: "POST method",
			args: commandArgs{
				path:   "/post-endpoint",
				method: "POST",
			},
			server: mockConfig{
				path: "/post-endpoint",
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write([]byte(r.Method)); err != nil { // #nosec
						t.Log(err)
					}
				},
			},
			wantStatus:   200,
			wantResponse: []byte("POST"),
		},
		{
			name: "PUT method",
			args: commandArgs{
				path:   "/put-endpoint",
				method: "PUT",
			},
			server: mockConfig{
				path: "/put-endpoint",
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write([]byte(r.Method)); err != nil { // #nosec
						t.Log(err)
					}
				},
			},
			wantStatus:   200,
			wantResponse: []byte("PUT"),
		},
		{
			name: "DELETE method",
			args: commandArgs{
				path:   "/delete-endpoint",
				method: "DELETE",
			},
			server: mockConfig{
				path: "/delete-endpoint",
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write([]byte(r.Method)); err != nil { // #nosec
						t.Log(err)
					}
				},
			},
			wantStatus:   200,
			wantResponse: []byte("DELETE"),
		},
		{
			name: "PATCH method",
			args: commandArgs{
				path:   "/patch-endpoint",
				method: "PATCH",
			},
			server: mockConfig{
				path: "/patch-endpoint",
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write([]byte(r.Method)); err != nil { // #nosec
						t.Log(err)
					}
				},
			},
			wantStatus:   200,
			wantResponse: []byte("PATCH"),
		},
		{
			name: "method case insensitive",
			args: commandArgs{
				path:   "/case-method",
				method: "post",
			},
			server: mockConfig{
				path: "/case-method",
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write([]byte(r.Method)); err != nil { // #nosec
						t.Log(err)
					}
				},
			},
			wantStatus:   200,
			wantResponse: []byte("POST"),
		},
		{
			name: "all headers are received",
			args: commandArgs{
				path: "/all-headers",
				headers: map[string]string{
					"X-Custom-1": "value1",
					"X-Custom-2": "value2",
				},
			},
			server: mockConfig{
				path: "/all-headers",
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					header1 := r.Header.Get("X-Custom-1")
					header2 := r.Header.Get("X-Custom-2")
					if _, err := w.Write([]byte(header1 + "-" + header2)); err != nil { // #nosec
						t.Log(err)
					}
				},
			},
			wantStatus:   200,
			wantResponse: []byte("value1-value2"),
		},
		{
			name: "response 201 created",
			args: commandArgs{
				path: "/created",
			},
			server: mockConfig{
				path:     "/created",
				status:   201,
				response: []byte(`{"id":123}`),
			},
			wantStatus:   201,
			wantResponse: []byte(`{"id":123}`),
		},
		{
			name: "response 204 no content",
			args: commandArgs{
				path: "/no-content",
			},
			server: mockConfig{
				path:     "/no-content",
				status:   204,
				response: []byte{},
			},
			wantStatus:   204,
			wantResponse: []byte{},
		},
		{
			name: "response 302 redirect is success",
			args: commandArgs{
				path: "/redirect",
			},
			server: mockConfig{
				path:     "/redirect",
				status:   http.StatusFound,
				response: []byte(""),
			},
			wantStatus:   http.StatusFound,
			wantResponse: []byte{},
		},
		{
			name: "JSON response content type",
			args: commandArgs{
				path: "/json-response",
			},
			server: mockConfig{
				path: "/json-response",
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write([]byte(`{"key":"value"}`)); err != nil { // #nosec
						t.Log(err)
					}
				},
			},
			wantStatus:   200,
			wantResponse: []byte(`{"key":"value"}`),
		},
		{
			name: "text response content type",
			args: commandArgs{
				path: "/text-response",
			},
			server: mockConfig{
				path: "/text-response",
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write([]byte("plain text response")); err != nil { // #nosec
						t.Log(err)
					}
				},
			},
			wantStatus:   200,
			wantResponse: []byte("plain text response"),
		},
		{
			name: "verbose flag enabled",
			args: commandArgs{
				path:    "/verbose-test",
				verbose: true,
			},
			server: mockConfig{
				path: "/verbose-test",
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write([]byte("response body")); err != nil { // #nosec
						t.Log(err)
					}
				},
			},
			wantStatus:   200,
			wantResponse: []byte("response body"),
		},
		{
			name: "empty response body",
			args: commandArgs{
				path: "/empty",
			},
			server: mockConfig{
				path:     "/empty",
				status:   200,
				response: []byte{},
			},
			wantStatus:   200,
			wantResponse: []byte{},
		},
		{
			name: "timeout respected for fast server",
			args: commandArgs{
				path:    "/timeout-ok",
				timeout: 5,
			},
			server: mockConfig{
				path:     "/timeout-ok",
				status:   200,
				response: []byte("OK"),
			},
			wantStatus:   200,
			wantResponse: []byte("OK"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerFunc := tt.server.handler
			if handlerFunc == nil {
				handlerFunc = func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, tt.server.path, r.URL.Path)
					w.WriteHeader(tt.server.status)
					if _, err := w.Write(tt.server.response); err != nil { // #nosec
						t.Log(err)
					}
				}
			}
			srv := httptest.NewServer(http.HandlerFunc(handlerFunc))
			defer srv.Close()

			serverDetails := &coreConfig.ServerDetails{
				Url:         srv.URL,
				AccessToken: "my-token",
			}

			ctx := newMockContext(&tt.args)

			var stdErr bytes.Buffer
			var stdOut bytes.Buffer

			prevLogger := clientlog.GetLogger()
			t.Cleanup(func() { clientlog.SetLogger(prevLogger) })
			clientlog.SetLogger(clientlog.NewLoggerWithFlags(clientlog.INFO, &stdErr, 0))

			err := runApiCmd(ctx, serverDetails, &stdOut)

			if tt.wantErr != nil {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.wantStatus > 0 {
				// Status is logged via client log (may include level prefix); body may be verbose on stderr too
				assert.Contains(t, stdErr.String(), fmt.Sprintf("Http Status: %d", tt.wantStatus))
			}

			if tt.wantResponse != nil {
				assert.Equal(t, strings.TrimSpace(string(tt.wantResponse)), strings.TrimSpace(stdOut.String()))
			}
		})
	}
}

func TestApiTimeoutExpired(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a server that takes longer than the client timeout.
		// The client will disconnect before the handler writes anything.
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	defer srv.Close()

	serverDetails := &coreConfig.ServerDetails{
		Url:         srv.URL,
		AccessToken: "my-token",
	}

	ctx := newMockContext(&commandArgs{
		path:    "/slow",
		timeout: 1, // 1-second timeout; server sleeps for 5 s
	})

	var stdOut bytes.Buffer
	err := runApiCmd(ctx, serverDetails, &stdOut)
	assert.Error(t, err, "expected a timeout error")
}

type commandArgs struct {
	path    string
	method  string
	headers map[string]string
	verbose bool
	timeout int
}

type mockContext struct {
	args      []string
	stringMap map[string]string
	boolMap   map[string]bool
	intMap    map[string]int
	setMap    map[string]bool
	sliceMap  map[string][]string
}

// NewMockContext creates a new mockContext with the given arguments
func newMockContext(cmdArgs *commandArgs) *mockContext {
	mc := &mockContext{
		stringMap: make(map[string]string),
		boolMap:   make(map[string]bool),
		intMap:    make(map[string]int),
		setMap:    make(map[string]bool),
		sliceMap:  make(map[string][]string),
	}
	if cmdArgs.path != "" {
		mc.args = append(mc.args, cmdArgs.path)
	}
	if cmdArgs.method != "" {
		mc.setString(flagMethod, cmdArgs.method)
	}
	if cmdArgs.headers != nil {
		var headers []string
		for k, v := range cmdArgs.headers {
			headers = append(headers, fmt.Sprintf("%v: %v", k, v))
		}
		mc.setStringSlice(flagHeader, headers)
	}
	if cmdArgs.verbose {
		mc.setBool(flagVerbose, true)
	}
	if cmdArgs.timeout > 0 {
		mc.setInt(flagTimeout, cmdArgs.timeout)
	}
	return mc
}

// Args returns the positional arguments
func (mc *mockContext) Args() cli.Args {
	return mc.args
}

// String returns the string value of a named flag
func (mc *mockContext) String(name string) string {
	if val, ok := mc.stringMap[name]; ok {
		return val
	}
	return ""
}

// Bool returns the boolean value of a named flag
func (mc *mockContext) Bool(name string) bool {
	if val, ok := mc.boolMap[name]; ok {
		return val
	}
	return false
}

// Int returns the integer value of a named flag
func (mc *mockContext) Int(name string) int {
	if val, ok := mc.intMap[name]; ok {
		return val
	}
	return 0
}

// IsSet returns whether a named flag was set
func (mc *mockContext) IsSet(name string) bool {
	if set, ok := mc.setMap[name]; ok {
		return set
	}
	return false
}

// StringSlice returns a string slice value of a named flag
func (mc *mockContext) StringSlice(name string) []string {
	if val, ok := mc.sliceMap[name]; ok {
		return val
	}
	return []string{}
}

// Helper methods for setting flag values in tests
func (mc *mockContext) setString(name, value string) {
	mc.stringMap[name] = value
	mc.setMap[name] = true
}

func (mc *mockContext) setBool(name string, value bool) {
	mc.boolMap[name] = value
	if value {
		mc.setMap[name] = true
	}
}

func (mc *mockContext) setStringSlice(name string, values []string) {
	mc.sliceMap[name] = values
	mc.setMap[name] = true
}

func (mc *mockContext) setInt(name string, value int) {
	mc.intMap[name] = value
	mc.setMap[name] = true
}
