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

func TestFlowPipelineGetCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		require.Equal(t, "/oapi/v1/flow/pipelines/p-1", r.URL.Path)
		require.Equal(t, "valid-token", r.Header.Get("x-yunxiao-token"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"p-1","name":"deploy-prod","status":"success"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "flow", "pipeline", "get", "--organization-id", "org-123", "--pipeline-id", "p-1", "--json")
	cmd.Env = testEnv(
		"YUNXIAO_ACCESS_TOKEN=valid-token",
		"YUNXIAO_API_BASE_URL="+server.URL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.True(t, called, "expected CLI to call Yunxiao Flow pipeline API")

	expected, err := os.ReadFile(filepath.Join(root, "test", "golden", "flow_pipeline_get_success.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(expected), stdout.String())
	require.Empty(t, stderr.String())
}

func TestFlowPipelinesListCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		require.Equal(t, "/oapi/v1/flow/pipelines", r.URL.Path)
		require.Equal(t, "valid-token", r.Header.Get("x-yunxiao-token"))
		require.Equal(t, "2", r.URL.Query().Get("perPage"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"p-1","name":"deploy-prod","status":"SUCCESS"},{"id":"p-2","name":"test","status":"RUNNING"}],"nextPage":"2"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "flow", "pipelines", "list", "--organization-id", "org-123", "--page-size", "2", "--json")
	cmd.Env = testEnv(
		"YUNXIAO_ACCESS_TOKEN=valid-token",
		"YUNXIAO_API_BASE_URL="+server.URL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.True(t, called, "expected CLI to call Yunxiao Flow pipelines API")

	expected, err := os.ReadFile(filepath.Join(root, "test", "golden", "flow_pipelines_list_success.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(expected), stdout.String())
	require.Empty(t, stderr.String())
}

func TestFlowPipelineGetNotFound(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"pipeline not found"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "flow", "pipeline", "get", "--organization-id", "org-123", "--pipeline-id", "nonexistent", "--json")
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
	require.Equal(t, 4, exitErr.ExitCode())

	expected, err := os.ReadFile(filepath.Join(root, "test", "golden", "flow_pipeline_get_not_found.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(expected), stdout.String())
	require.Contains(t, stderr.String(), "pipeline lookup failed")
}

func TestFlowPipelineGetQuietSuppressesRetryDiagnostics(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `{"message":"service unavailable"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "flow", "pipeline", "get", "--organization-id", "org-123", "--pipeline-id", "p-1", "--json", "--quiet")
	cmd.Env = testEnv(
		"YUNXIAO_ACCESS_TOKEN=valid-token",
		"YUNXIAO_API_BASE_URL="+server.URL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	require.NotContains(t, stderr.String(), "[RETRY]")
	require.Contains(t, stderr.String(), "pipeline lookup failed")
}

func TestFlowPipelineGetRateLimitedRetriesRealHTTP(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"message":"rate limited"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "flow", "pipeline", "get", "--organization-id", "org-123", "--pipeline-id", "p-1", "--json")
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
	require.Equal(t, 5, exitErr.ExitCode())
	require.Equal(t, 4, attempts)

	expected, err := os.ReadFile(filepath.Join(root, "test", "golden", "flow_pipeline_get_rate_limited.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(expected), stdout.String())
	require.Contains(t, stderr.String(), "attempt 1/3")
	require.Contains(t, stderr.String(), "attempt 2/3")
	require.Contains(t, stderr.String(), "attempt 3/3")
}

func TestFlowPipelineGetUpstream500RetriesRealHTTP(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"message":"server error"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "flow", "pipeline", "get", "--organization-id", "org-123", "--pipeline-id", "p-1", "--json")
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
	require.Equal(t, 7, exitErr.ExitCode())
	require.Equal(t, 4, attempts)

	expected, err := os.ReadFile(filepath.Join(root, "test", "golden", "flow_pipeline_get_upstream_500.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(expected), stdout.String())
}

func TestFlowPipelineGetUpstream503RetriesRealHTTP(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `{"message":"service unavailable"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "flow", "pipeline", "get", "--organization-id", "org-123", "--pipeline-id", "p-1", "--json")
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
	require.Equal(t, 7, exitErr.ExitCode())
	require.Equal(t, 4, attempts)

	expected, err := os.ReadFile(filepath.Join(root, "test", "golden", "flow_pipeline_get_upstream_503.json"))
	require.NoError(t, err)
	require.JSONEq(t, string(expected), stdout.String())
	require.Contains(t, stderr.String(), "attempt 1/3")
	require.Contains(t, stderr.String(), "attempt 2/3")
	require.Contains(t, stderr.String(), "attempt 3/3")
}
