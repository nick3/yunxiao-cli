package integration

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func buildTestBinary(t *testing.T, root string) string {
	t.Helper()
	absRoot, err := filepath.Abs(root)
	require.NoError(t, err)

	binary := filepath.Join(absRoot, "yunxiao-test")
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/yunxiao")
	cmd.Dir = absRoot
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	t.Cleanup(func() { _ = os.Remove(binary) })
	return binary
}

func testEnv(overrides ...string) []string {
	base := make([]string, 0, len(os.Environ())+len(overrides))
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "YUNXIAO_") {
			continue
		}
		base = append(base, env)
	}
	return append(base, overrides...)
}

func TestOrgCurrentAuthFailure(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "org", "current", "--json")
	cmd.Env = testEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)

	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 3, exitErr.ExitCode())

	expected, err := os.ReadFile(filepath.Join(root, "test", "golden", "org_current_auth_failed.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(expected), stdout.String())
	require.Empty(t, stderr.String())
}

func TestOrgCurrentSuccess(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/oapi/v1/platform/user", r.URL.Path)
		require.Equal(t, "valid-token", r.Header.Get("x-yunxiao-token"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"user-001","name":"agent-user","organization":{"id":"org-123","name":"demo-org"}}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "org", "current", "--json", "--trace-id", "trace-success")
	cmd.Env = testEnv(
		"YUNXIAO_ACCESS_TOKEN=valid-token",
		"YUNXIAO_API_BASE_URL="+server.URL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)

	expected, err := os.ReadFile(filepath.Join(root, "test", "golden", "org_current_success.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(expected), stdout.String())
	require.Empty(t, stderr.String())
}

func TestOrgCurrentPrefersEnvTokenOverConfigToken(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("access_token: config-token\n"), 0o600))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/oapi/v1/platform/user", r.URL.Path)
		require.Equal(t, "env-token", r.Header.Get("x-yunxiao-token"))
		fmt.Fprint(w, `{"id":"user-1"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "org", "current", "--json")
	cmd.Env = testEnv(
		"YUNXIAO_CONFIG_FILE="+configDir,
		"YUNXIAO_ACCESS_TOKEN=env-token",
		"YUNXIAO_API_BASE_URL="+server.URL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"id": "user-1"`)
	require.Empty(t, stderr.String())
}

func TestOrgCurrentDecodeFailure(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `not-json`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "org", "current", "--json")
	cmd.Env = testEnv(
		"YUNXIAO_ACCESS_TOKEN=valid-token",
		"YUNXIAO_API_BASE_URL="+server.URL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 1, exitErr.ExitCode())

	expected, err := os.ReadFile(filepath.Join(root, "test", "golden", "org_current_decode_failed.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(expected), stdout.String())
	require.Contains(t, stderr.String(), "response decode failed")
}

func TestOrgCurrentForbidden(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"permission denied"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "org", "current", "--json")
	cmd.Env = testEnv(
		"YUNXIAO_ACCESS_TOKEN=valid-token",
		"YUNXIAO_API_BASE_URL="+server.URL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 8, exitErr.ExitCode())

	expected, err := os.ReadFile(filepath.Join(root, "test", "golden", "org_current_forbidden.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(expected), stdout.String())
	require.Contains(t, stderr.String(), "permission denied")
}

func TestOrgCurrentTimeout(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id":"slow-user"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "org", "current", "--json", "--timeout", "1")
	cmd.Env = testEnv(
		"YUNXIAO_ACCESS_TOKEN=valid-token",
		"YUNXIAO_API_BASE_URL="+server.URL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 6, exitErr.ExitCode())
	expected, err := os.ReadFile(filepath.Join(root, "test", "golden", "org_current_timeout.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(expected), stdout.String())
}
