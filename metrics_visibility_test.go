package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
)

type capturedRequest struct {
	Method string
	Path   string
	Body   []byte
}

type visReq struct {
	Method string
	Path   string
	Body   []byte
}

func startMockServer(t *testing.T) (*httptest.Server, chan capturedRequest) {
	t.Helper()
	ch := make(chan capturedRequest, 4)
	handler := http.NewServeMux()
	handler.HandleFunc("/artifactory/api/system/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"version":"7.200.0"}`))
	})
	handler.HandleFunc("/artifactory/api/system/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	// Call Home stub
	handler.HandleFunc("/artifactory/api/system/usage", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	handler.HandleFunc("/jfconnect/api/v1/backoffice/metrics/log", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, _ := io.ReadAll(r.Body)
		ch <- capturedRequest{Method: r.Method, Path: r.URL.Path, Body: body}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})
	srv := httptest.NewServer(handler)
	return srv, ch
}

func startVisMockServer(t *testing.T) (*httptest.Server, chan visReq) {
	t.Helper()
	ch := make(chan visReq, 4)
	mux := http.NewServeMux()
	mux.HandleFunc("/artifactory/api/system/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"version":"7.200.0"}`))
	})
	mux.HandleFunc("/artifactory/api/system/usage", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	mux.HandleFunc("/jfconnect/api/v1/backoffice/metrics/log", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		b, _ := io.ReadAll(r.Body)
		ch <- visReq{Method: r.Method, Path: r.URL.Path, Body: b}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})
	srv := httptest.NewServer(mux)
	return srv, ch
}

func sortCSV(csv string) string {
	if csv == "" {
		return ""
	}
	parts := strings.Split(csv, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func TestVisibilitySendUsage_RtCurl_E2E(t *testing.T) {
	srv, ch := startMockServer(t)
	defer srv.Close()

	// Isolate CLI home and enable usage reporting
	home := t.TempDir()
	_ = os.Setenv("JFROG_CLI_HOME_DIR", home)
	_ = os.Setenv("JFROG_CLI_REPORT_USAGE", "true")

	jf := coreTests.NewJfrogCli(execMain, "jf", "").WithoutCredentials()

	// Create mock server config pointing to the httptest server
	platformURL := srv.URL + "/"
	artURL := srv.URL + "/artifactory/"
	if err := jf.Exec(
		"c", "add", "mock",
		"--url", platformURL,
		"--artifactory-url", artURL,
		"--access-token", "dummy",
		"--interactive=false",
		"--enc-password=false",
	); err != nil {
		t.Fatalf("config add failed: %v", err)
	}
	if err := jf.Exec("c", "use", "mock"); err != nil {
		t.Fatalf("config use failed: %v", err)
	}

	// Run real CLI command pointing to mock Artifactory via server-id
	if err := jf.Exec("rt", "curl", "-X", "POST", "/api/system/ping", "--server-id", "mock"); err != nil {
		t.Fatalf("jf exec failed: %v", err)
	}

	// Assert metric was posted
	select {
	case req := <-ch:
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.Path != "/jfconnect/api/v1/backoffice/metrics/log" {
			t.Fatalf("unexpected path: %s", req.Path)
		}
		var payload struct {
			Name   string `json:"metrics_name"`
			Labels struct {
				Flags     string `json:"flags"`
				FeatureID string `json:"feature_id"`
			} `json:"labels"`
		}
		if err := json.Unmarshal(req.Body, &payload); err != nil {
			t.Fatalf("bad JSON: %v", err)
		}
		if payload.Name != "jfcli_commands_count" {
			t.Fatalf("unexpected metric name: %s", payload.Name)
		}
		if payload.Labels.FeatureID != "rt_curl" {
			t.Fatalf("unexpected feature_id: %s", payload.Labels.FeatureID)
		}
		// rt curl removes flags internally; expect empty
		if payload.Labels.Flags != "" {
			t.Fatalf("expected empty flags string, got %q", payload.Labels.Flags)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout waiting for metrics POST")
	}
}

func TestVisibilitySendUsage_RtPing_Flags(t *testing.T) {
	srv, ch := startMockServer(t)
	defer srv.Close()

	// Isolate CLI home and enable usage reporting
	home := t.TempDir()
	_ = os.Setenv("JFROG_CLI_HOME_DIR", home)
	_ = os.Setenv("JFROG_CLI_REPORT_USAGE", "true")

	jf := coreTests.NewJfrogCli(execMain, "jf", "").WithoutCredentials()

	// Create mock server config pointing to the httptest server
	platformURL := srv.URL + "/"
	artURL := srv.URL + "/artifactory/"
	if err := jf.Exec(
		"c", "add", "mock",
		"--url", platformURL,
		"--artifactory-url", artURL,
		"--access-token", "dummy",
		"--interactive=false",
		"--enc-password=false",
	); err != nil {
		t.Fatalf("config add failed: %v", err)
	}
	if err := jf.Exec("c", "use", "mock"); err != nil {
		t.Fatalf("config use failed: %v", err)
	}

	// Run ping with a flag to assert capture
	if err := jf.Exec("rt", "ping", "--server-id", "mock"); err != nil {
		// ping may fail in some contexts; metrics should still be sent
	}

	// Assert metric was posted with expected flags
	select {
	case req := <-ch:
		if req.Path != "/jfconnect/api/v1/backoffice/metrics/log" {
			t.Fatalf("unexpected path: %s", req.Path)
		}
		var payload struct {
			Labels struct {
				Flags string `json:"flags"`
			} `json:"labels"`
		}
		if err := json.Unmarshal(req.Body, &payload); err != nil {
			t.Fatalf("bad JSON: %v", err)
		}
		if payload.Labels.Flags != "server-id" {
			t.Fatalf("flags mismatch: got %q want %q", payload.Labels.Flags, "server-id")
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout waiting for metrics POST")
	}
}

func TestVisibility_GoBuild_Flags(t *testing.T) {
	srv, ch := startVisMockServer(t)
	defer srv.Close()

	// Isolated home and enable usage
	home := t.TempDir()
	_ = os.Setenv("JFROG_CLI_HOME_DIR", home)
	_ = os.Setenv("JFROG_CLI_REPORT_USAGE", "true")

	jf := coreTests.NewJfrogCli(execMain, "jf", "").WithoutCredentials()

	// Configure mock platform
	platformURL := srv.URL + "/"
	artURL := srv.URL + "/artifactory/"
	if err := jf.Exec("c", "add", "mock", "--url", platformURL, "--artifactory-url", artURL, "--access-token", "dummy", "--interactive=false", "--enc-password=false"); err != nil {
		t.Fatalf("config add failed: %v", err)
	}
	if err := jf.Exec("c", "use", "mock"); err != nil {
		t.Fatalf("config use failed: %v", err)
	}

	// Create a minimal Go project
	projDir := t.TempDir()
	goMod := []byte("module example.com/jfvis\n\ngo 1.21\n")
	mainGo := []byte("package main\nfunc main(){}\n")
	if err := os.WriteFile(projDir+"/go.mod", goMod, 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(projDir+"/main.go", mainGo, 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	// Provide a minimal buildtools config so GoCmd passes verification
	if err := os.MkdirAll(projDir+"/.jfrog/projects", 0o755); err != nil {
		t.Fatalf("mkdir .jfrog/projects: %v", err)
	}
	goYaml := []byte("version: 1\n" +
		"type: go\n" +
		"resolver:\n" +
		"    repo: go-virtual\n" +
		"    serverId: mock\n" +
		"deployer:\n" +
		"    repo: go-virtual\n" +
		"    serverId: mock\n")
	if err := os.WriteFile(projDir+"/.jfrog/projects/go.yaml", goYaml, 0o644); err != nil {
		t.Fatalf("write go.yaml: %v", err)
	}
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	_ = os.Chdir(projDir)

	// Run a buildtools command (go build) with flags; ignore execution error
	_ = jf.Exec("go", "build", "--build-name", "test", "--build-number", "1", "--server-id", "mock")

	select {
	case req := <-ch:
		if req.Path != "/jfconnect/api/v1/backoffice/metrics/log" {
			t.Fatalf("unexpected path: %s", req.Path)
		}
		var p struct {
			Labels struct {
				Flags string `json:"flags"`
			} `json:"labels"`
		}
		if err := json.Unmarshal(req.Body, &p); err != nil {
			t.Fatalf("bad JSON: %v", err)
		}
		if p.Labels.Flags != "build-name,build-number,server-id" {
			t.Fatalf("flags mismatch: got %q want %q", p.Labels.Flags, "build-name,build-number,server-id")
		}
	case <-time.After(15 * time.Second):
		t.Fatal("timeout waiting for metric")
	}
}
