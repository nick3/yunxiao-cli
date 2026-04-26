package integration

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodeupReposListRequiresOrganizationID(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "codeup", "repos", "list", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"code": "PARAM_REQUIRED"`)
	require.Contains(t, stdout.String(), `"category": "param"`)
	require.Empty(t, stderr.String())
}

func TestCodeupReposListCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		require.Equal(t, "/oapi/v1/codeup/repositories", r.URL.Path)
		require.Equal(t, "valid-token", r.Header.Get("x-yunxiao-token"))
		require.Equal(t, "2", r.URL.Query().Get("perPage"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"repo-1","name":"frontend"},{"id":"repo-2","name":"backend"}],"nextPage":"abc123"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "codeup", "repos", "list", "--organization-id", "org-123", "--page-size", "2", "--json")
	cmd.Env = testEnv(
		"YUNXIAO_ACCESS_TOKEN=valid-token",
		"YUNXIAO_API_BASE_URL="+server.URL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.True(t, called, "expected CLI to call Yunxiao Codeup repositories API")

	expected, err := os.ReadFile(filepath.Join(root, "test", "golden", "codeup_repos_list_success.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(expected), stdout.String())
	require.Empty(t, stderr.String())
}

func TestCodeupRepoGetEmptyResponseFails(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "codeup", "repo", "get", "--organization-id", "org-123", "--repo-id", "repo-empty", "--json")
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

	expected, err := os.ReadFile(filepath.Join(root, "test", "golden", "codeup_repo_get_empty_response.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(expected), stdout.String())
	require.Contains(t, stderr.String(), "repository lookup failed")
}

func TestCodeupRepoGetCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		require.Equal(t, "/oapi/v1/codeup/repositories/repo-1", r.URL.Path)
		require.Equal(t, "valid-token", r.Header.Get("x-yunxiao-token"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"repo-1","name":"frontend","path":"frontend"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "codeup", "repo", "get", "--organization-id", "org-123", "--repo-id", "repo-1", "--json")
	cmd.Env = testEnv(
		"YUNXIAO_ACCESS_TOKEN=valid-token",
		"YUNXIAO_API_BASE_URL="+server.URL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.True(t, called, "expected CLI to call Yunxiao Codeup repository API")

	expected, err := os.ReadFile(filepath.Join(root, "test", "golden", "codeup_repo_get_success.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(expected), stdout.String())
	require.Empty(t, stderr.String())
}
