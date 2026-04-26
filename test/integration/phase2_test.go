package integration

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjexProjectsListCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/oapi/v1/projex/projects:search", r.URL.Path)
		fmt.Fprint(w, `{"data":[{"id":"proj-1","name":"demo"}],"nextPage":"2"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "projects", "list", "--organization-id", "org-123", "--page-size", "1", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":"proj-1","name":"demo"}],"meta":{"pagination":{"next_token":"2","page_size":1,"has_more":true}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexProjectGetCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/projex/projects/proj-1", r.URL.Path)
		fmt.Fprint(w, `{"id":"proj-1","name":"demo"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "project", "get", "--organization-id", "org-123", "--project-id", "proj-1", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"id":"proj-1","name":"demo"},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestPackagesArtifactsListCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/packages/repositories/repo-1/artifacts", r.URL.Path)
		require.Equal(t, "maven", r.URL.Query().Get("repoType"))
		require.Equal(t, "2", r.URL.Query().Get("perPage"))
		w.Header().Set("x-next-page", "3")
		fmt.Fprint(w, `[{"id":101,"name":"artifact-a"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "packages", "artifacts", "list", "--organization-id", "org-123", "--repo-id", "repo-1", "--repo-type", "maven", "--page-size", "2", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":101,"name":"artifact-a"}],"meta":{"pagination":{"next_token":"3","page_size":2,"has_more":true}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestPackagesReposListUsesNextPageHeader(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/packages/repositories", r.URL.Path)
		w.Header().Set("x-next-page", "4")
		fmt.Fprint(w, `[{"id":"pkg-repo-1","name":"repo"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "packages", "repos", "list", "--organization-id", "org-123", "--page-size", "2", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":"pkg-repo-1","name":"repo"}],"meta":{"pagination":{"next_token":"4","page_size":2,"has_more":true}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestPackagesArtifactGetCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/packages/repositories/repo-1/artifacts/101", r.URL.Path)
		require.Equal(t, "maven", r.URL.Query().Get("repoType"))
		fmt.Fprint(w, `{"id":101,"name":"artifact-a"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "packages", "artifact", "get", "--organization-id", "org-123", "--repo-id", "repo-1", "--artifact-id", "101", "--repo-type", "maven", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"id":101,"name":"artifact-a"},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestTesthubTestcasesListCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/oapi/v1/testhub/testRepos/repo-1/testcases:search", r.URL.Path)
		fmt.Fprint(w, `{"data":[{"id":"tc-1","name":"login"}],"nextPage":null}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "testhub", "testcases", "list", "--organization-id", "org-123", "--test-repo-id", "repo-1", "--page-size", "2", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":"tc-1","name":"login"}],"meta":{"pagination":{"next_token":null,"page_size":2,"has_more":false}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestTesthubTestcaseGetCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/testhub/testRepos/repo-1/testcases/tc-1", r.URL.Path)
		fmt.Fprint(w, `{"id":"tc-1","name":"login"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "testhub", "testcase", "get", "--organization-id", "org-123", "--test-repo-id", "repo-1", "--testcase-id", "tc-1", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"id":"tc-1","name":"login"},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestRawRequestGetCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/custom/resource", r.URL.Path)
		require.Equal(t, "1", r.URL.Query().Get("foo"))
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "raw", "request", "--method", "GET", "--path", "/oapi/v1/custom/resource?foo=1", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"ok":true},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestRawRequestRejectsNonReadOnlyMethod(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "raw", "request", "--method", "POST", "--path", "/oapi/v1/custom/resource", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"code": "PARAM_INVALID"`)
	require.Empty(t, stderr.String())
}

func TestRawRequestRejectsInvalidMethodBeforeAuth(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "raw", "request", "--method", "POST", "--path", "/oapi/v1/custom/resource", "--json")
	cmd.Env = testEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"code": "PARAM_INVALID"`)
	require.NotContains(t, stdout.String(), `"code": "AUTH_FAILED"`)
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemsListCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/oapi/v1/projex/workitems:search", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"category":"Task","spaceId":"proj-1","perPage":2}`, string(body))
		fmt.Fprint(w, `{"data":[{"id":"wi-1","subject":"fix bug"}],"nextPage":"3"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitems", "list", "--organization-id", "org-123", "--category", "Task", "--space-id", "proj-1", "--page-size", "2", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":"wi-1","subject":"fix bug"}],"meta":{"pagination":{"next_token":"3","page_size":2,"has_more":true}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemGetCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/projex/workitems/wi-1", r.URL.Path)
		fmt.Fprint(w, `{"id":"wi-1","subject":"fix bug"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem", "get", "--organization-id", "org-123", "--workitem-id", "wi-1", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"id":"wi-1","subject":"fix bug"},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexSprintsListCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/projex/projects/proj-1/sprints", r.URL.Path)
		require.Equal(t, "2", r.URL.Query().Get("perPage"))
		w.Header().Set("x-next-page", "4")
		fmt.Fprint(w, `[{"id":"sprint-1","name":"Sprint 1"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "sprints", "list", "--organization-id", "org-123", "--project-id", "proj-1", "--page-size", "2", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":"sprint-1","name":"Sprint 1"}],"meta":{"pagination":{"next_token":"4","page_size":2,"has_more":true}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestTesthubDirectoriesListCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/testhub/testRepos/repo-1/directories", r.URL.Path)
		fmt.Fprint(w, `{"directories":[{"id":"dir-1","name":"smoke"}]}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "testhub", "directories", "list", "--organization-id", "org-123", "--test-repo-id", "repo-1", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"directories":[{"id":"dir-1","name":"smoke"}]},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestTesthubTestplansListCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/oapi/v1/projex/testPlan/list", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{}`, string(body))
		fmt.Fprint(w, `[{"id":"plan-1","name":"regression"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "testhub", "testplans", "list", "--organization-id", "org-123", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":"plan-1","name":"regression"}],"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestListCommandsRejectInvalidPageSizeBeforeAuth(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cases := [][]string{
		{"codeup", "repos", "list", "--organization-id", "org-123", "--page-size", "0", "--json"},
		{"flow", "pipelines", "list", "--organization-id", "org-123", "--page-size", "0", "--json"},
		{"projex", "projects", "list", "--organization-id", "org-123", "--page-size", "0", "--json"},
		{"projex", "workitems", "list", "--organization-id", "org-123", "--category", "Task", "--space-id", "proj-1", "--page-size", "0", "--json"},
		{"projex", "sprints", "list", "--organization-id", "org-123", "--project-id", "proj-1", "--page-size", "0", "--json"},
		{"packages", "repos", "list", "--organization-id", "org-123", "--page-size", "0", "--json"},
		{"packages", "artifacts", "list", "--organization-id", "org-123", "--repo-id", "repo-1", "--repo-type", "maven", "--page-size", "0", "--json"},
		{"testhub", "testcases", "list", "--organization-id", "org-123", "--test-repo-id", "repo-1", "--page-size", "0", "--json"},
	}

	for _, args := range cases {
		cmd := exec.Command(binary, args...)
		cmd.Env = testEnv()

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		require.Error(t, err, args)
		exitErr, ok := err.(*exec.ExitError)
		require.True(t, ok, args)
		require.Equal(t, 2, exitErr.ExitCode(), args)
		require.Contains(t, stdout.String(), `"code": "PARAM_INVALID"`, args)
		require.Contains(t, stdout.String(), "page_size must be greater than 0", args)
		require.NotContains(t, stdout.String(), `"code": "AUTH_FAILED"`, args)
		require.Empty(t, stderr.String(), args)
	}
}

func TestListCommandsRejectNonIntegerPageSizeBeforeAuth(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cases := [][]string{
		{"codeup", "repos", "list", "--organization-id", "org-123", "--page-size", "abc", "--json"},
		{"flow", "pipelines", "list", "--organization-id", "org-123", "--page-size", "abc", "--json"},
		{"projex", "projects", "list", "--organization-id", "org-123", "--page-size", "abc", "--json"},
		{"projex", "workitems", "list", "--organization-id", "org-123", "--category", "Task", "--space-id", "proj-1", "--page-size", "abc", "--json"},
		{"projex", "sprints", "list", "--organization-id", "org-123", "--project-id", "proj-1", "--page-size", "abc", "--json"},
		{"packages", "repos", "list", "--organization-id", "org-123", "--page-size", "abc", "--json"},
		{"packages", "artifacts", "list", "--organization-id", "org-123", "--repo-id", "repo-1", "--repo-type", "maven", "--page-size", "abc", "--json"},
		{"testhub", "testcases", "list", "--organization-id", "org-123", "--test-repo-id", "repo-1", "--page-size", "abc", "--json"},
	}

	for _, args := range cases {
		cmd := exec.Command(binary, args...)
		cmd.Env = testEnv()

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		require.Error(t, err, args)
		exitErr, ok := err.(*exec.ExitError)
		require.True(t, ok, args)
		require.Equal(t, 2, exitErr.ExitCode(), args)
		require.Contains(t, stdout.String(), `"code": "PARAM_INVALID"`, args)
		require.Contains(t, stdout.String(), `invalid argument \"abc\" for \"--page-size\" flag`, args)
		require.NotContains(t, stdout.String(), `"code": "AUTH_FAILED"`, args)
		require.NotContains(t, stdout.String(), `"code": "COMMAND_FAILED"`, args)
	}
}
