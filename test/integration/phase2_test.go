package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjexProjectsListCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/oapi/v1/projex/projects:search", r.URL.Path)
		w.Header().Set("x-next-page", "2")
		w.Header().Set("x-per-page", "5")
		fmt.Fprint(w, `[{"id":"proj-1","name":"demo"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "projects", "list", "--organization-id", "org-123", "--page-size", "1", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":"proj-1","name":"demo"}],"meta":{"pagination":{"next_token":"2","page_size":5,"has_more":true}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexProjectsListAcceptsWrapperSearchResponse(t *testing.T) {
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

func TestProjexProjectsListMineResolvesCurrentUserAndOrganization(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/platform/user":
			require.Equal(t, http.MethodGet, r.Method)
			fmt.Fprint(w, `{"id":"user-1","name":"Nick","lastOrganization":"org-123"}`)
		case "/oapi/v1/projex/organizations/org-123/projects:search":
			require.Equal(t, http.MethodPost, r.Method)
			var payload map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			require.Equal(t, float64(2), payload["perPage"])
			var extraConditions struct {
				ConditionGroups [][]struct {
					ClassName       string   `json:"className"`
					FieldIdentifier string   `json:"fieldIdentifier"`
					Format          string   `json:"format"`
					Operator        string   `json:"operator"`
					Value           []string `json:"value"`
				} `json:"conditionGroups"`
			}
			require.NoError(t, json.Unmarshal([]byte(payload["extraConditions"].(string)), &extraConditions))
			require.Len(t, extraConditions.ConditionGroups, 1)
			require.Len(t, extraConditions.ConditionGroups[0], 1)
			condition := extraConditions.ConditionGroups[0][0]
			require.Equal(t, "user", condition.ClassName)
			require.Equal(t, "users", condition.FieldIdentifier)
			require.Equal(t, "multiList", condition.Format)
			require.Equal(t, "CONTAINS", condition.Operator)
			require.Equal(t, []string{"user-1"}, condition.Value)
			fmt.Fprint(w, `[{"id":"proj-1","name":"demo"}]`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "projects", "list", "--mine", "--page-size", "2", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+strings.Replace(server.URL, "http://", "http://openapi-rdc.aliyuncs.com@", 1))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":"proj-1","name":"demo"}],"meta":{"pagination":{"next_token":null,"page_size":2,"has_more":false}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexProjectsListMineRejectsMissingCurrentUserID(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/oapi/v1/platform/user", r.URL.Path)
		fmt.Fprint(w, `{"lastOrganization":"org-123"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "projects", "list", "--mine", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+strings.Replace(server.URL, "http://", "http://openapi-rdc.aliyuncs.com@", 1))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"code": "PARAM_REQUIRED"`)
	require.Contains(t, stdout.String(), "current user response has no id")
	require.Empty(t, stderr.String())
}

func TestProjexProjectsListRejectsMineWithScenarioFilter(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "projex", "projects", "list", "--mine", "--scenario-filter", "manage", "--json")
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
	require.Contains(t, stdout.String(), "mine cannot be used")
	require.Empty(t, stderr.String())
}

func TestProjexProjectsListRejectsUserIDWithoutScenarioFilter(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "projex", "projects", "list", "--user-id", "user-1", "--json")
	cmd.Env = testEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"code": "PARAM_REQUIRED"`)
	require.Contains(t, stdout.String(), "scenario_filter is required")
	require.Empty(t, stderr.String())
}

func TestProjexProjectsListSendsMCPCompatibleFilters(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/oapi/v1/projex/projects:search", r.URL.Path)

		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, float64(10), payload["perPage"])
		require.Equal(t, float64(3), payload["page"])
		require.Equal(t, "name", payload["orderBy"])
		require.Equal(t, "asc", payload["sort"])

		var conditions struct {
			ConditionGroups [][]struct {
				ClassName       string   `json:"className"`
				FieldIdentifier string   `json:"fieldIdentifier"`
				Format          string   `json:"format"`
				Operator        string   `json:"operator"`
				ToValue         *string  `json:"toValue"`
				Value           []string `json:"value"`
			} `json:"conditionGroups"`
		}
		require.NoError(t, json.Unmarshal([]byte(payload["conditions"].(string)), &conditions))
		require.Len(t, conditions.ConditionGroups, 1)
		fields := map[string][]string{}
		for _, condition := range conditions.ConditionGroups[0] {
			fields[condition.FieldIdentifier] = condition.Value
		}
		require.Equal(t, []string{"demo"}, fields["name"])
		require.Equal(t, []string{"active"}, fields["status"])
		require.Equal(t, []string{"creator-1"}, fields["creator"])
		require.Equal(t, []string{"admin-1"}, fields["project.admin"])
		require.Equal(t, []string{"normal"}, fields["logicalStatus"])

		fmt.Fprint(w, `[{"id":"proj-1","name":"demo"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "projects", "list", "--organization-id", "org-123", "--page-size", "10", "--page-token", "3", "--name", "demo", "--status", "active", "--created-after", "2026-04-01", "--created-before", "2026-04-26", "--creator", "creator-1", "--admin-user-id", "admin-1", "--logical-status", "normal", "--order-by", "name", "--sort", "asc", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":"proj-1","name":"demo"}],"meta":{"pagination":{"next_token":null,"page_size":10,"has_more":false}},"error":null}`, stdout.String())
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

func TestProjexProjectTemplatesListDecodesArrayAndWrappedArray(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	tests := []struct {
		name string
		body string
	}{
		{name: "array", body: `[{"id":"tmpl-1","name":"Basic"}]`},
		{name: "wrapped result", body: `{"result":[{"id":"tmpl-1","name":"Basic"}]}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/oapi/v1/projex/organizations/org-123/projectTemplates", r.URL.Path)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()

			cmd := exec.Command(binary, "projex", "project-templates", "list", "--organization-id", "org-123", "--json")
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+strings.Replace(server.URL, "http://", "http://openapi-rdc.aliyuncs.com@", 1))

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.NoError(t, err)
			require.JSONEq(t, `{"version":"v1","data":[{"id":"tmpl-1","name":"Basic"}],"meta":{},"error":null}`, stdout.String())
			require.Empty(t, stderr.String())
		})
	}
}

func TestProjexProjectTemplateFieldsDecodesObjectAndWrappedObject(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	tests := []struct {
		name string
		body string
	}{
		{name: "object", body: `{"fields":[{"identifier":"priority"}]}`},
		{name: "wrapped result", body: `{"result":{"fields":[{"identifier":"priority"}]}}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/oapi/v1/projex/organizations/org-123/projectTemplates/tmpl-1/fields", r.URL.Path)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()

			cmd := exec.Command(binary, "projex", "project-template", "fields", "--organization-id", "org-123", "--template-id", "tmpl-1", "--json")
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+strings.Replace(server.URL, "http://", "http://openapi-rdc.aliyuncs.com@", 1))

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.NoError(t, err)
			require.JSONEq(t, `{"version":"v1","data":{"fields":[{"identifier":"priority"}]},"meta":{},"error":null}`, stdout.String())
			require.Empty(t, stderr.String())
		})
	}
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
		w.Header().Set("x-page", "1")
		w.Header().Set("x-next-page", "3")
		w.Header().Set("x-prev-page", "1")
		w.Header().Set("x-total-pages", "8")
		w.Header().Set("x-total", "15")
		fmt.Fprint(w, `[{"id":"wi-1","subject":"fix bug"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitems", "list", "--organization-id", "org-123", "--category", "Task", "--space-id", "proj-1", "--page-size", "2", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":"wi-1","subject":"fix bug"}],"meta":{"pagination":{"next_token":"3","page_size":2,"has_more":true,"page":1,"total_pages":8,"total":15,"prev_token":"1"}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemsListAcceptsProjectIDAlias(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/oapi/v1/projex/workitems:search", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"category":"Task","spaceId":"proj-1","perPage":20}`, string(body))
		fmt.Fprint(w, `[{"id":"wi-1"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitems", "list", "--organization-id", "org-123", "--category", "Task", "--project-id", "proj-1", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"error": null`)
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemsListRejectsMismatchedProjectAndSpaceIDBeforeAuth(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "projex", "workitems", "list", "--organization-id", "org-123", "--category", "Task", "--project-id", "proj-1", "--space-id", "proj-2", "--json")
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
	require.Contains(t, stdout.String(), "project_id and space_id must match")
	require.NotContains(t, stdout.String(), `"code": "AUTH_FAILED"`)
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemsListRejectsInvalidSearchObjectResponses(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	for _, tc := range []struct {
		name     string
		body     string
		code     string
		message  string
		exitCode int
	}{
		{name: "unexpected object", body: `{"unexpected":true}`, code: "RESPONSE_DECODE_FAILED", message: "expected object with data array", exitCode: 1},
		{name: "business error", body: `{"success":false,"errorMessage":"search failed"}`, code: "UPSTREAM_BUSINESS_ERROR", message: "search failed", exitCode: 7},
		{name: "code message", body: `{"code":"InvalidParameter","message":"bad query"}`, code: "UPSTREAM_BUSINESS_ERROR", message: "bad query", exitCode: 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/oapi/v1/projex/workitems:search", r.URL.Path)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()

			cmd := exec.Command(binary, "projex", "workitems", "list", "--organization-id", "org-123", "--category", "Task", "--space-id", "proj-1", "--json")
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.Error(t, err)
			exitErr, ok := err.(*exec.ExitError)
			require.True(t, ok)
			require.Equal(t, tc.exitCode, exitErr.ExitCode())
			require.Contains(t, stdout.String(), `"code": "`+tc.code+`"`)
			require.Contains(t, stdout.String(), `"data": null`)
			require.Contains(t, stderr.String(), tc.message)
		})
	}
}

func TestProjexWorkitemsListMineRejectsInvalidFlagCombinations(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	tests := []struct {
		name    string
		args    []string
		message string
	}{
		{name: "unfinished without mine", args: []string{"projex", "workitems", "list", "--unfinished", "--category", "Task", "--json"}, message: "unfinished can only be used with mine"},
		{name: "mine with space id", args: []string{"projex", "workitems", "list", "--mine", "--category", "Task", "--space-id", "proj-1", "--json"}, message: "project_id or space_id cannot be used with mine"},
		{name: "mine with project id", args: []string{"projex", "workitems", "list", "--mine", "--category", "Task", "--project-id", "proj-1", "--json"}, message: "project_id or space_id cannot be used with mine"},
		{name: "mine with page token", args: []string{"projex", "workitems", "list", "--mine", "--category", "Task", "--page-token", "2", "--json"}, message: "page_token cannot be used with mine"},
		{name: "mine with assigned to", args: []string{"projex", "workitems", "list", "--mine", "--category", "Task", "--assigned-to", "user-2", "--json"}, message: "assigned_to cannot be used with mine"},
		{name: "mine with advanced conditions", args: []string{"projex", "workitems", "list", "--mine", "--category", "Task", "--advanced-conditions", `{"conditionGroups":[]}`, "--json"}, message: "advanced_conditions cannot be used with mine"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, tc.args...)
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
			require.Contains(t, stdout.String(), tc.message)
			require.NotContains(t, stdout.String(), `"code": "AUTH_FAILED"`)
			require.Empty(t, stderr.String())
		})
	}
}

func TestProjexWorkitemsListMineAggregatesParticipatedProjects(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	workitemPages := map[string]int{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/platform/user":
			require.Equal(t, http.MethodGet, r.Method)
			fmt.Fprint(w, `{"id":"user-1","lastOrganization":"org-123"}`)
		case "/oapi/v1/projex/organizations/org-123/projects:search":
			require.Equal(t, http.MethodPost, r.Method)
			var payload map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			require.Equal(t, float64(2), payload["perPage"])
			var extraConditions struct {
				ConditionGroups [][]struct {
					FieldIdentifier string   `json:"fieldIdentifier"`
					Value           []string `json:"value"`
				} `json:"conditionGroups"`
			}
			require.NoError(t, json.Unmarshal([]byte(payload["extraConditions"].(string)), &extraConditions))
			require.Equal(t, "users", extraConditions.ConditionGroups[0][0].FieldIdentifier)
			require.Equal(t, []string{"user-1"}, extraConditions.ConditionGroups[0][0].Value)
			w.Header().Set("x-total", "2")
			fmt.Fprint(w, `[{"id":"proj-1"},{"id":"proj-2"}]`)
		case "/oapi/v1/projex/organizations/org-123/workitems:search":
			require.Equal(t, http.MethodPost, r.Method)
			var payload map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			require.Equal(t, "Task", payload["category"])
			require.Equal(t, float64(2), payload["perPage"])
			var conditions struct {
				ConditionGroups [][]struct {
					FieldIdentifier string   `json:"fieldIdentifier"`
					Value           []string `json:"value"`
				} `json:"conditionGroups"`
			}
			require.NoError(t, json.Unmarshal([]byte(payload["conditions"].(string)), &conditions))
			require.Equal(t, "assignedTo", conditions.ConditionGroups[0][0].FieldIdentifier)
			require.Equal(t, []string{"user-1"}, conditions.ConditionGroups[0][0].Value)

			spaceID := payload["spaceId"].(string)
			workitemPages[spaceID]++
			switch {
			case spaceID == "proj-1" && workitemPages[spaceID] == 1:
				w.Header().Set("x-next-page", "2")
				w.Header().Set("x-total", "3")
				fmt.Fprint(w, `[{"id":"wi-open","status":{"name":"待处理"}},{"id":"wi-done","status":{"name":"已完成"}}]`)
			case spaceID == "proj-1" && workitemPages[spaceID] == 2:
				require.Equal(t, float64(2), payload["page"])
				fmt.Fprint(w, `[{"id":"wi-closed","logicalStatus":"COMPLETED"}]`)
			case spaceID == "proj-2":
				fmt.Fprint(w, `[{"id":"wi-other","status":{"id":"done-marker","name":"未完成"}}]`)
			default:
				t.Fatalf("unexpected workitem request for %s page %d", spaceID, workitemPages[spaceID])
			}
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitems", "list", "--mine", "--unfinished", "--category", "Task", "--page-size", "2", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+strings.Replace(server.URL, "http://", "http://openapi-rdc.aliyuncs.com@", 1))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":"wi-open","status":{"name":"待处理"}},{"id":"wi-other","status":{"id":"done-marker","name":"未完成"}}],"meta":{"pagination":{"next_token":null,"page_size":2,"has_more":false,"total":2}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemsListMineFailsFastOnIncompleteAggregate(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	run := func(t *testing.T, handler http.HandlerFunc) (string, string, int) {
		t.Helper()
		server := httptest.NewServer(handler)
		defer server.Close()

		cmd := exec.Command(binary, "projex", "workitems", "list", "--mine", "--unfinished", "--category", "Task", "--page-size", "2", "--json")
		cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+strings.Replace(server.URL, "http://", "http://openapi-rdc.aliyuncs.com@", 1))

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		require.Error(t, err)
		exitErr, ok := err.(*exec.ExitError)
		require.True(t, ok, "unexpected command error type %T: %v; stdout=%s; stderr=%s", err, err, stdout.String(), stderr.String())
		return stdout.String(), stderr.String(), exitErr.ExitCode()
	}

	t.Run("missing project id", func(t *testing.T) {
		stdout, stderr, exitCode := run(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/oapi/v1/platform/user":
				fmt.Fprint(w, `{"id":"user-1","lastOrganization":"org-123"}`)
			case "/oapi/v1/projex/organizations/org-123/projects:search":
				fmt.Fprint(w, `[{"name":"missing-id"}]`)
			default:
				http.Error(w, "unexpected path", http.StatusInternalServerError)
			}
		})
		require.Equal(t, 1, exitCode)
		require.Contains(t, stdout, `"code": "RESPONSE_DECODE_FAILED"`)
		require.Contains(t, stdout, `"data": null`)
		require.Contains(t, stderr, "project list response item 0")
	})

	t.Run("repeated page token", func(t *testing.T) {
		workitemRequests := 0
		stdout, stderr, exitCode := run(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/oapi/v1/platform/user":
				fmt.Fprint(w, `{"id":"user-1","lastOrganization":"org-123"}`)
			case "/oapi/v1/projex/organizations/org-123/projects:search":
				fmt.Fprint(w, `[{"id":"proj-1"}]`)
			case "/oapi/v1/projex/organizations/org-123/workitems:search":
				workitemRequests++
				w.Header().Set("x-next-page", "2")
				fmt.Fprint(w, `[{"id":"wi-open","status":{"name":"待处理"}}]`)
			default:
				http.Error(w, "unexpected path", http.StatusInternalServerError)
			}
		})
		require.Equal(t, 1, exitCode)
		require.Equal(t, 2, workitemRequests)
		require.Contains(t, stdout, `"code": "PAGINATION_LOOP_DETECTED"`)
		require.Contains(t, stdout, `"data": null`)
		require.Contains(t, stderr, "repeated next page token")
	})

	t.Run("project pagination total without next token", func(t *testing.T) {
		stdout, stderr, exitCode := run(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/oapi/v1/platform/user":
				fmt.Fprint(w, `{"id":"user-1","lastOrganization":"org-123"}`)
			case "/oapi/v1/projex/organizations/org-123/projects:search":
				w.Header().Set("x-total", "2")
				fmt.Fprint(w, `[{"id":"proj-1"}]`)
			default:
				http.Error(w, "unexpected path", http.StatusInternalServerError)
			}
		})
		require.Equal(t, 1, exitCode)
		require.Contains(t, stdout, `"code": "PAGINATION_INVALID"`)
		require.Contains(t, stdout, `"data": null`)
		require.Contains(t, stderr, "participated project search returned 1 of 2 items")
	})

	t.Run("workitem pagination total without next token", func(t *testing.T) {
		stdout, stderr, exitCode := run(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/oapi/v1/platform/user":
				fmt.Fprint(w, `{"id":"user-1","lastOrganization":"org-123"}`)
			case "/oapi/v1/projex/organizations/org-123/projects:search":
				fmt.Fprint(w, `[{"id":"proj-1"}]`)
			case "/oapi/v1/projex/organizations/org-123/workitems:search":
				w.Header().Set("x-total", "2")
				fmt.Fprint(w, `[{"id":"wi-open","status":{"name":"待处理"}}]`)
			default:
				http.Error(w, "unexpected path", http.StatusInternalServerError)
			}
		})
		require.Equal(t, 1, exitCode)
		require.Contains(t, stdout, `"code": "PAGINATION_INVALID"`)
		require.Contains(t, stdout, `"data": null`)
		require.Contains(t, stderr, "workitem search for project proj-1 returned 1 of 2 items")
	})

	t.Run("unclassified status", func(t *testing.T) {
		stdout, stderr, exitCode := run(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/oapi/v1/platform/user":
				fmt.Fprint(w, `{"id":"user-1","lastOrganization":"org-123"}`)
			case "/oapi/v1/projex/organizations/org-123/projects:search":
				fmt.Fprint(w, `[{"id":"proj-1"}]`)
			case "/oapi/v1/projex/organizations/org-123/workitems:search":
				fmt.Fprint(w, `[{"id":"wi-unknown","status":{"name":"QA Review"}}]`)
			default:
				http.Error(w, "unexpected path", http.StatusInternalServerError)
			}
		})
		require.Equal(t, 1, exitCode)
		require.Contains(t, stdout, `"code": "WORKITEM_STATUS_UNCLASSIFIED"`)
		require.Contains(t, stdout, `"data": null`)
		require.Contains(t, stderr, "wi-unknown")
		require.Contains(t, stderr, "QA Review")
	})

	t.Run("conflicting status fields", func(t *testing.T) {
		stdout, stderr, exitCode := run(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/oapi/v1/platform/user":
				fmt.Fprint(w, `{"id":"user-1","lastOrganization":"org-123"}`)
			case "/oapi/v1/projex/organizations/org-123/projects:search":
				fmt.Fprint(w, `[{"id":"proj-1"}]`)
			case "/oapi/v1/projex/organizations/org-123/workitems:search":
				fmt.Fprint(w, `[{"id":"wi-conflict","status":{"name":"已完成","stage":"open"}}]`)
			default:
				http.Error(w, "unexpected path", http.StatusInternalServerError)
			}
		})
		require.Equal(t, 1, exitCode)
		require.Contains(t, stdout, `"code": "WORKITEM_STATUS_UNCLASSIFIED"`)
		require.Contains(t, stdout, `"data": null`)
		require.Contains(t, stderr, "wi-conflict")
		require.Contains(t, stderr, "已完成")
		require.Contains(t, stderr, "open")
	})

	t.Run("conflicting logical status and status", func(t *testing.T) {
		stdout, stderr, exitCode := run(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/oapi/v1/platform/user":
				fmt.Fprint(w, `{"id":"user-1","lastOrganization":"org-123"}`)
			case "/oapi/v1/projex/organizations/org-123/projects:search":
				fmt.Fprint(w, `[{"id":"proj-1"}]`)
			case "/oapi/v1/projex/organizations/org-123/workitems:search":
				fmt.Fprint(w, `[{"id":"wi-cross-conflict","logicalStatus":"COMPLETED","status":{"name":"待处理"}}]`)
			default:
				http.Error(w, "unexpected path", http.StatusInternalServerError)
			}
		})
		require.Equal(t, 1, exitCode)
		require.Contains(t, stdout, `"code": "WORKITEM_STATUS_UNCLASSIFIED"`)
		require.Contains(t, stdout, `"data": null`)
		require.Contains(t, stderr, "wi-cross-conflict")
		require.Contains(t, stderr, "COMPLETED")
		require.Contains(t, stderr, "待处理")
	})

	t.Run("project workitem failure", func(t *testing.T) {
		stdout, stderr, exitCode := run(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/oapi/v1/platform/user":
				fmt.Fprint(w, `{"id":"user-1","lastOrganization":"org-123"}`)
			case "/oapi/v1/projex/organizations/org-123/projects:search":
				fmt.Fprint(w, `[{"id":"proj-1"},{"id":"proj-2"}]`)
			case "/oapi/v1/projex/organizations/org-123/workitems:search":
				var payload map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
				if payload["spaceId"] == "proj-2" {
					w.WriteHeader(http.StatusForbidden)
					fmt.Fprint(w, `{"message":"no permission"}`)
					return
				}
				fmt.Fprint(w, `[{"id":"wi-open","status":{"name":"待处理"}}]`)
			default:
				http.Error(w, "unexpected path", http.StatusInternalServerError)
			}
		})
		require.Equal(t, 8, exitCode)
		require.Contains(t, stdout, `"data": null`)
		require.Contains(t, stdout, "proj-2")
		require.NotContains(t, stdout, "wi-open")
		require.Contains(t, stderr, "failed to list assigned workitems for project proj-2")
	})
}

func TestProjexWorkitemsListHelpExposesRequiredFlags(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "projex", "workitems", "list", "--help", "--json")
	cmd.Env = testEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.Empty(t, stderr.String())

	var envelope struct {
		Data struct {
			Flags []struct {
				Name string `json:"name"`
			} `json:"flags"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &envelope))

	flagNames := make(map[string]bool, len(envelope.Data.Flags))
	for _, flag := range envelope.Data.Flags {
		flagNames[flag.Name] = true
	}
	require.True(t, flagNames["organization-id"])
	require.True(t, flagNames["category"])
	require.True(t, flagNames["space-id"])
	require.True(t, flagNames["mine"])
	require.True(t, flagNames["unfinished"])
	require.True(t, flagNames["page-size"])
	require.True(t, flagNames["page-token"])
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

func TestProjexSprintsListRejectsInvalidPaginationHeaders(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/oapi/v1/projex/projects/proj-1/sprints", r.URL.Path)
		w.Header().Set("x-total", "invalid")
		fmt.Fprint(w, `[{"id":"sprint-1","name":"Sprint 1"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "sprints", "list", "--organization-id", "org-123", "--project-id", "proj-1", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 1, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"code": "PAGINATION_INVALID"`)
	require.Contains(t, stdout.String(), `"data": null`)
	require.Contains(t, stderr.String(), `x-total="invalid"`)
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

func TestProjexWorkitemCreateCallsRegionYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/oapi/v1/projex/workitems", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"spaceId":"proj-1","workitemTypeId":"type-1","subject":"fix bug","assignedTo":"user-1","customFieldValues":{"customText":"cf text"}}`, string(body))
		fmt.Fprint(w, `{"id":"wi-1","subject":"fix bug"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem", "create", "--organization-id", "org-123", "--space-id", "proj-1", "--workitem-type-id", "type-1", "--subject", "fix bug", "--assigned-to", "user-1", "--custom-field", "customText=cf text", "--yes", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"id":"wi-1","subject":"fix bug"},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemCreateAllowsReservedCustomFieldKeys(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/oapi/v1/projex/workitems", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"spaceId":"proj-1","workitemTypeId":"type-1","subject":"fix bug","assignedTo":"user-1","customFieldValues":{"subject":"custom subject","assignedTo":"custom assignee"}}`, string(body))
		fmt.Fprint(w, `{"id":"wi-1","subject":"fix bug"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem", "create", "--organization-id", "org-123", "--space-id", "proj-1", "--workitem-type-id", "type-1", "--subject", "fix bug", "--assigned-to", "user-1", "--custom-field", "subject=custom subject", "--custom-fields-json", `{"assignedTo":"custom assignee"}`, "--yes", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"id":"wi-1","subject":"fix bug"},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemCreateAcceptsProjectIDAliasAndAssignedToSelf(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/platform/user":
			require.Equal(t, http.MethodGet, r.Method)
			fmt.Fprint(w, `{"id":"user-self"}`)
		case "/oapi/v1/projex/workitems":
			require.Equal(t, http.MethodPost, r.Method)
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.JSONEq(t, `{"spaceId":"proj-1","workitemTypeId":"type-1","subject":"fix bug","assignedTo":"user-self"}`, string(body))
			fmt.Fprint(w, `{"id":"wi-1"}`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem", "create", "--organization-id", "org-123", "--project-id", "proj-1", "--workitem-type-id", "type-1", "--subject", "fix bug", "--assigned-to", "self", "--yes", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"id":"wi-1"},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemCreateRejectsMismatchedProjectAndSpaceIDBeforeAuth(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "projex", "workitem", "create", "--organization-id", "org-123", "--project-id", "proj-1", "--space-id", "proj-2", "--workitem-type-id", "type-1", "--subject", "fix bug", "--assigned-to", "user-1", "--yes", "--json")
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
	require.Contains(t, stdout.String(), "project_id and space_id must match")
	require.NotContains(t, stdout.String(), `"code": "AUTH_FAILED"`)
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemCreateReadsDescriptionFile(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	descriptionPath := filepath.Join(t.TempDir(), "description.md")
	require.NoError(t, os.WriteFile(descriptionPath, []byte("file description"), 0o600))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "file description", payload["description"])
		fmt.Fprint(w, `{"id":"wi-1"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem", "create", "--organization-id", "org-123", "--space-id", "proj-1", "--workitem-type-id", "type-1", "--subject", "fix bug", "--assigned-to", "user-1", "--description-file", descriptionPath, "--yes", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"error": null`)
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemCreateWithoutYesDoesNotSendRequest(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	requests := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem", "create", "--organization-id", "org-123", "--space-id", "proj-1", "--workitem-type-id", "type-1", "--subject", "fix bug", "--assigned-to", "user-1", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), "yes")
	require.NotContains(t, stdout.String(), `"code": "AUTH_FAILED"`)
	require.Empty(t, stderr.String())
	require.Equal(t, 0, requests)
}

func TestProjexWorkitemCreateRejectsAckOnlyResponses(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "success only", body: `{"success":true}`},
		{name: "empty result", body: `{"success":true,"result":{}}`},
		{name: "null result", body: `{"success":true,"result":null}`},
		{name: "array result", body: `{"success":true,"result":[]}`},
		{name: "unexpected object", body: `{"success":true,"result":{"unexpected":true}}`},
		{name: "queued object", body: `{"result":{"queued":true}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/oapi/v1/projex/workitems", r.URL.Path)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()

			cmd := exec.Command(binary, "projex", "workitem", "create", "--organization-id", "org-123", "--space-id", "proj-1", "--workitem-type-id", "type-1", "--subject", "fix bug", "--assigned-to", "user-1", "--yes", "--json")
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.Error(t, err)
			exitErr, ok := err.(*exec.ExitError)
			require.True(t, ok)
			require.Equal(t, 1, exitErr.ExitCode())
			require.Contains(t, stdout.String(), `"code": "RESPONSE_DECODE_FAILED"`)
			require.Contains(t, stdout.String(), `"data": null`)
			require.Contains(t, stderr.String(), "failed to decode")
		})
	}
}

func TestProjexWorkitemDescriptionFileRejectsUnsafeInputs(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	dir := t.TempDir()
	emptyPath := filepath.Join(dir, "empty.md")
	invalidUTF8Path := filepath.Join(dir, "invalid.md")
	largePath := filepath.Join(dir, "large.md")
	symlinkPath := filepath.Join(dir, "link.md")
	targetPath := filepath.Join(dir, "target.md")
	require.NoError(t, os.WriteFile(emptyPath, nil, 0o600))
	require.NoError(t, os.WriteFile(invalidUTF8Path, []byte{0xff}, 0o600))
	require.NoError(t, os.WriteFile(largePath, bytes.Repeat([]byte("a"), 1024*1024+1), 0o600))
	require.NoError(t, os.WriteFile(targetPath, []byte("target"), 0o600))
	require.NoError(t, os.Symlink(targetPath, symlinkPath))

	for _, tc := range []struct {
		name    string
		path    string
		message string
	}{
		{name: "empty", path: emptyPath, message: "file is empty"},
		{name: "invalid utf8", path: invalidUTF8Path, message: "file must be UTF-8"},
		{name: "too large", path: largePath, message: "file exceeds 1MiB"},
		{name: "symlink", path: symlinkPath, message: "file must be a regular file"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, "projex", "workitem", "create", "--organization-id", "org-123", "--space-id", "proj-1", "--workitem-type-id", "type-1", "--subject", "fix bug", "--assigned-to", "user-1", "--description-file", tc.path, "--yes", "--json")
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token")

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.Error(t, err)
			exitErr, ok := err.(*exec.ExitError)
			require.True(t, ok)
			require.Equal(t, 1, exitErr.ExitCode())
			require.Contains(t, stdout.String(), `"code": "FILE_READ_FAILED"`)
			require.Contains(t, stdout.String(), tc.message)
			require.Contains(t, stdout.String(), filepath.Base(tc.path))
			require.NotContains(t, stdout.String(), dir)
			require.NotContains(t, stdout.String(), `"code": "AUTH_FAILED"`)
			require.Empty(t, stderr.String())
		})
	}
}

func TestProjexWorkitemUpdateSendsStandardAndCustomFieldsAtTopLevel(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method)
		require.Equal(t, "/oapi/v1/projex/workitems/wi-1", r.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "new subject", payload["subject"])
		require.Equal(t, "user-1", payload["assignedTo"])
		require.Equal(t, "high", payload["priority"])
		require.Equal(t, "cf text", payload["customText"])
		require.Equal(t, "42", payload["customNumber"])
		require.NotContains(t, payload, "customFieldValues")
		fmt.Fprint(w, `{"id":"wi-1","subject":"new subject"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem", "update", "--organization-id", "org-123", "--workitem-id", "wi-1", "--subject", "new subject", "--assigned-to", "user-1", "--priority", "high", "--custom-field", "customText=cf text", "--custom-field", "customNumber=42", "--yes", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"workitem_id":"wi-1","updated":true},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemUpdateRejectsReservedCustomFieldKeysBeforeAuth(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	requests := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	}))
	defer server.Close()

	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "pair", args: []string{"--custom-field", "subject=override"}},
		{name: "json", args: []string{"--custom-fields-json", `{"assignedTo":"user-2"}`}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{"projex", "workitem", "update", "--organization-id", "org-123", "--workitem-id", "wi-1", "--subject", "new subject", "--yes", "--json"}
			args = append(args, tc.args...)
			cmd := exec.Command(binary, args...)
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.Error(t, err)
			exitErr, ok := err.(*exec.ExitError)
			require.True(t, ok)
			require.Equal(t, 2, exitErr.ExitCode())
			require.Contains(t, stdout.String(), `"code": "PARAM_INVALID"`)
			require.Contains(t, stdout.String(), "conflicts with a standard workitem field")
			require.NotContains(t, stdout.String(), `"code": "AUTH_FAILED"`)
			require.Empty(t, stderr.String())
		})
	}
	require.Equal(t, 0, requests)
}

func TestProjexWorkitemUpdateResolvesAssignedToSelf(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oapi/v1/platform/user":
			require.Equal(t, http.MethodGet, r.Method)
			fmt.Fprint(w, `{"id":"user-self"}`)
		case "/oapi/v1/projex/workitems/wi-1":
			require.Equal(t, http.MethodPut, r.Method)
			var payload map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			require.Equal(t, "user-self", payload["assignedTo"])
			fmt.Fprint(w, `{"success":true}`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem", "update", "--organization-id", "org-123", "--workitem-id", "wi-1", "--assigned-to", "self", "--yes", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"workitem_id":"wi-1","updated":true},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemUpdateAcceptsExplicitSuccessConfirmation(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "success only", body: `{"success":true}`},
		{name: "success with request id", body: `{"success":true,"requestId":"req-1"}`},
		{name: "success with uppercase request id", body: `{"success":true,"RequestId":"req-1"}`},
		{name: "success with request ID", body: `{"success":true,"requestID":"req-1"}`},
		{name: "success with trace id", body: `{"success":true,"traceId":"trace-1"}`},
		{name: "success with trace ID", body: `{"success":true,"traceID":"trace-1"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(t, "/oapi/v1/projex/workitems/wi-1", r.URL.Path)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()

			cmd := exec.Command(binary, "projex", "workitem", "update", "--organization-id", "org-123", "--workitem-id", "wi-1", "--subject", "new subject", "--yes", "--json")
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.NoError(t, err)
			require.JSONEq(t, `{"version":"v1","data":{"workitem_id":"wi-1","updated":true},"meta":{},"error":null}`, stdout.String())
			require.Empty(t, stderr.String())
		})
	}
}

func TestProjexWorkitemUpdateRejectsAmbiguousSuccessResponses(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "code only", body: `{"code":"InvalidParameter"}`},
		{name: "unexpected object", body: `{"unexpected":true}`},
		{name: "success with empty result", body: `{"success":true,"result":{}}`},
		{name: "success with unexpected metadata", body: `{"success":true,"unexpected":"value"}`},
		{name: "success with unexpected result", body: `{"success":true,"result":{"unexpected":true}}`},
		{name: "queued result", body: `{"result":{"queued":true}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(t, "/oapi/v1/projex/workitems/wi-1", r.URL.Path)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()

			cmd := exec.Command(binary, "projex", "workitem", "update", "--organization-id", "org-123", "--workitem-id", "wi-1", "--subject", "new subject", "--yes", "--json")
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.Error(t, err)
			exitErr, ok := err.(*exec.ExitError)
			require.True(t, ok)
			require.Equal(t, 1, exitErr.ExitCode())
			require.Contains(t, stdout.String(), `"code": "RESPONSE_DECODE_FAILED"`)
			require.Contains(t, stdout.String(), `"data": null`)
			require.Contains(t, stderr.String(), "expected resource object with id or explicit success confirmation")
		})
	}
}

func TestProjexWorkitemUpdateRejectsNoUpdateFieldsBeforeAuth(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "projex", "workitem", "update", "--organization-id", "org-123", "--workitem-id", "wi-1", "--yes", "--json")
	cmd.Env = testEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"code": "PARAM_REQUIRED"`)
	require.Contains(t, stdout.String(), "update")
	require.NotContains(t, stdout.String(), `"code": "AUTH_FAILED"`)
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemUpdateRejectsEmptySubjectBeforeAuth(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	requests := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem", "update", "--organization-id", "org-123", "--workitem-id", "wi-1", "--subject", "", "--yes", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"code": "PARAM_REQUIRED"`)
	require.Contains(t, stdout.String(), "update")
	require.NotContains(t, stdout.String(), `"code": "AUTH_FAILED"`)
	require.Empty(t, stderr.String())
	require.Equal(t, 0, requests)
}

func TestProjexWorkitemUpdateWithoutYesDoesNotSendRequest(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	requests := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem", "update", "--organization-id", "org-123", "--workitem-id", "wi-1", "--subject", "new subject", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), "yes")
	require.NotContains(t, stdout.String(), `"code": "AUTH_FAILED"`)
	require.Empty(t, stderr.String())
	require.Equal(t, 0, requests)
}

func TestProjexWorkitemTypesListAllCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/projex/workitemTypes", r.URL.Path)
		fmt.Fprint(w, `[{"id":"type-1","name":"Task"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem-types", "list", "--organization-id", "org-123", "--all", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":"type-1","name":"Task"}],"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemTypesListProjectCategoryCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/projex/projects/proj-1/workitemTypes", r.URL.Path)
		require.Equal(t, "Task", r.URL.Query().Get("category"))
		fmt.Fprint(w, `[{"id":"type-1","name":"Task"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem-types", "list", "--organization-id", "org-123", "--project-id", "proj-1", "--category", "Task", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":"type-1","name":"Task"}],"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemTypeMetadataCommandsCallYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	tests := []struct {
		name string
		args []string
		path string
		body string
	}{
		{name: "get", args: []string{"projex", "workitem-type", "get", "--organization-id", "org-123", "--workitem-type-id", "type-1"}, path: "/oapi/v1/projex/workitemTypes/type-1", body: `{"id":"type-1","name":"Task"}`},
		{name: "relations", args: []string{"projex", "workitem-types", "relations", "--organization-id", "org-123", "--workitem-type-id", "type-1"}, path: "/oapi/v1/projex/workitemTypes/type-1/relationWorkitemTypes", body: `[{"id":"rel-1"}]`},
		{name: "fields", args: []string{"projex", "workitem-type", "fields", "--organization-id", "org-123", "--project-id", "proj-1", "--workitem-type-id", "type-1"}, path: "/oapi/v1/projex/projects/proj-1/workitemTypes/type-1/fields", body: `{"fields":[{"identifier":"subject"}]}`},
		{name: "workflow", args: []string{"projex", "workitem-type", "workflow", "--organization-id", "org-123", "--project-id", "proj-1", "--workitem-type-id", "type-1"}, path: "/oapi/v1/projex/projects/proj-1/workitemTypes/type-1/workflows", body: `{"statuses":[{"id":"open"}]}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, tc.path, r.URL.Path)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()

			args := append(append([]string{}, tc.args...), "--json")
			cmd := exec.Command(binary, args...)
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.NoError(t, err)
			require.Contains(t, stdout.String(), `"version": "v1"`)
			require.Empty(t, stderr.String())
		})
	}
}

func TestProjexCommentsListArrayWithoutPaginationHeadersHasNoMorePages(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/projex/workitems/wi-1/comments", r.URL.Path)
		require.Equal(t, "2", r.URL.Query().Get("page"))
		require.Equal(t, "5", r.URL.Query().Get("perPage"))
		fmt.Fprint(w, `[{"id":"comment-1","content":"hello"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem", "comments", "list", "--organization-id", "org-123", "--workitem-id", "wi-1", "--page-token", "2", "--page-size", "5", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":[{"id":"comment-1","content":"hello"}],"meta":{"pagination":{"next_token":null,"page_size":5,"has_more":false}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexCommentsListRejectsInvalidPaginationHeader(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	for _, tc := range []struct {
		name   string
		header string
		value  string
	}{
		{name: "total", header: "x-total", value: "invalid"},
		{name: "per page", header: "x-per-page", value: "invalid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/oapi/v1/projex/workitems/wi-1/comments", r.URL.Path)
				w.Header().Set(tc.header, tc.value)
				fmt.Fprint(w, `[{"id":"comment-1","content":"hello"}]`)
			}))
			defer server.Close()

			cmd := exec.Command(binary, "projex", "workitem", "comments", "list", "--organization-id", "org-123", "--workitem-id", "wi-1", "--json")
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.Error(t, err)
			exitErr, ok := err.(*exec.ExitError)
			require.True(t, ok)
			require.Equal(t, 1, exitErr.ExitCode())
			require.Contains(t, stdout.String(), `"code": "PAGINATION_INVALID"`)
			require.Contains(t, stdout.String(), `"data": null`)
			require.Contains(t, stderr.String(), tc.header+`="`+tc.value+`"`)
		})
	}
}

func TestProjexListCommandsRejectUnexpectedObjectResponse(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	for _, tc := range []struct {
		name     string
		args     []string
		path     string
		resource string
	}{
		{name: "comments", args: []string{"projex", "workitem", "comments", "list", "--organization-id", "org-123", "--workitem-id", "wi-1", "--json"}, path: "/oapi/v1/projex/workitems/wi-1/comments", resource: "workitem comments"},
		{name: "workitem types", args: []string{"projex", "workitem-types", "list", "--organization-id", "org-123", "--all", "--json"}, path: "/oapi/v1/projex/workitemTypes", resource: "workitem metadata"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, tc.path, r.URL.Path)
				fmt.Fprint(w, `{"unexpected":true}`)
			}))
			defer server.Close()

			cmd := exec.Command(binary, tc.args...)
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.Error(t, err)
			exitErr, ok := err.(*exec.ExitError)
			require.True(t, ok)
			require.Equal(t, 1, exitErr.ExitCode())
			require.Contains(t, stdout.String(), `"code": "RESPONSE_DECODE_FAILED"`)
			require.Contains(t, stdout.String(), `"data": null`)
			require.Contains(t, stderr.String(), "failed to decode "+tc.resource+" response")
			require.Contains(t, stderr.String(), "expected array or object with result array")
		})
	}
}

func TestProjexListCommandsRejectBusinessErrorEnvelope(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	for _, tc := range []struct {
		name string
		args []string
		path string
	}{
		{name: "comments", args: []string{"projex", "workitem", "comments", "list", "--organization-id", "org-123", "--workitem-id", "wi-1", "--json"}, path: "/oapi/v1/projex/workitems/wi-1/comments"},
		{name: "workitem types", args: []string{"projex", "workitem-types", "list", "--organization-id", "org-123", "--all", "--json"}, path: "/oapi/v1/projex/workitemTypes"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, tc.path, r.URL.Path)
				fmt.Fprint(w, `{"success":false,"errorMessage":"metadata unavailable"}`)
			}))
			defer server.Close()

			cmd := exec.Command(binary, tc.args...)
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.Error(t, err)
			exitErr, ok := err.(*exec.ExitError)
			require.True(t, ok)
			require.Equal(t, 7, exitErr.ExitCode())
			require.Contains(t, stdout.String(), `"code": "UPSTREAM_BUSINESS_ERROR"`)
			require.Contains(t, stdout.String(), "metadata unavailable")
			require.Contains(t, stderr.String(), "metadata unavailable")
		})
	}
}

func TestProjexCommentCreateCallsRegionYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/oapi/v1/projex/workitems/wi-1/comments", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"content":"hello"}`, string(body))
		fmt.Fprint(w, `{"id":"comment-1","content":"hello"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem", "comment", "create", "--organization-id", "org-123", "--workitem-id", "wi-1", "--content", "hello", "--yes", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"id":"comment-1","content":"hello"},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexCommentCreateReadsContentFile(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	contentPath := filepath.Join(t.TempDir(), "comment.txt")
	require.NoError(t, os.WriteFile(contentPath, []byte("file comment"), 0o600))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"content":"file comment"}`, string(body))
		fmt.Fprint(w, `{"id":"comment-1","content":"file comment"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem", "comment", "create", "--organization-id", "org-123", "--workitem-id", "wi-1", "--content-file", contentPath, "--yes", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"id":"comment-1","content":"file comment"},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexCommentCreateRejectsAckOnlyResponses(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "success only", body: `{"success":true}`},
		{name: "unexpected object", body: `{"success":true,"result":{"unexpected":true}}`},
		{name: "queued object", body: `{"result":{"queued":true}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/oapi/v1/projex/workitems/wi-1/comments", r.URL.Path)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()

			cmd := exec.Command(binary, "projex", "workitem", "comment", "create", "--organization-id", "org-123", "--workitem-id", "wi-1", "--content", "hello", "--yes", "--json")
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.Error(t, err)
			exitErr, ok := err.(*exec.ExitError)
			require.True(t, ok)
			require.Equal(t, 1, exitErr.ExitCode())
			require.Contains(t, stdout.String(), `"code": "RESPONSE_DECODE_FAILED"`)
			require.Contains(t, stdout.String(), `"data": null`)
			require.Contains(t, stderr.String(), "failed to decode")
		})
	}
}

func TestProjexWriteCommandsRejectBusinessErrorEnvelope(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	for _, tc := range []struct {
		name string
		args []string
		path string
		body string
	}{
		{name: "create", args: []string{"projex", "workitem", "create", "--organization-id", "org-123", "--space-id", "proj-1", "--workitem-type-id", "type-1", "--subject", "fix bug", "--assigned-to", "user-1", "--yes", "--json"}, path: "/oapi/v1/projex/workitems", body: `{"success":false,"errorCode":"FIELD_INVALID","errorMessage":"field invalid"}`},
		{name: "update", args: []string{"projex", "workitem", "update", "--organization-id", "org-123", "--workitem-id", "wi-1", "--subject", "new subject", "--yes", "--json"}, path: "/oapi/v1/projex/workitems/wi-1", body: `{"code":"InvalidParameter","message":"field invalid"}`},
		{name: "comment create", args: []string{"projex", "workitem", "comment", "create", "--organization-id", "org-123", "--workitem-id", "wi-1", "--content", "hello", "--yes", "--json"}, path: "/oapi/v1/projex/workitems/wi-1/comments", body: `{"Code":"InvalidParameter","Message":"field invalid"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tc.path, r.URL.Path)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()

			cmd := exec.Command(binary, tc.args...)
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.Error(t, err)
			exitErr, ok := err.(*exec.ExitError)
			require.True(t, ok)
			require.Equal(t, 7, exitErr.ExitCode())
			require.Contains(t, stdout.String(), `"code": "UPSTREAM_BUSINESS_ERROR"`)
			require.Contains(t, stdout.String(), "field invalid")
			require.Contains(t, stderr.String(), "field invalid")
		})
	}
}

func TestProjexWriteCommandsDoNotRetryServerErrors(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	for _, tc := range []struct {
		name string
		args []string
		path string
	}{
		{name: "create", args: []string{"projex", "workitem", "create", "--organization-id", "org-123", "--space-id", "proj-1", "--workitem-type-id", "type-1", "--subject", "fix bug", "--assigned-to", "user-1", "--yes", "--json"}, path: "/oapi/v1/projex/workitems"},
		{name: "update", args: []string{"projex", "workitem", "update", "--organization-id", "org-123", "--workitem-id", "wi-1", "--subject", "new subject", "--yes", "--json"}, path: "/oapi/v1/projex/workitems/wi-1"},
		{name: "comment create", args: []string{"projex", "workitem", "comment", "create", "--organization-id", "org-123", "--workitem-id", "wi-1", "--content", "hello", "--yes", "--json"}, path: "/oapi/v1/projex/workitems/wi-1/comments"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				require.Equal(t, tc.path, r.URL.Path)
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprint(w, `{"message":"try later"}`)
			}))
			defer server.Close()

			cmd := exec.Command(binary, tc.args...)
			cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

			var stdout bytes.Buffer
			cmd.Stdout = &stdout

			err := cmd.Run()
			require.Error(t, err)
			exitErr, ok := err.(*exec.ExitError)
			require.True(t, ok)
			require.Equal(t, 7, exitErr.ExitCode())
			require.Contains(t, stdout.String(), `"code": "UPSTREAM_UNAVAILABLE"`)
			require.Equal(t, 1, requests)
		})
	}
}

func TestProjexCommentCreateWithoutYesDoesNotSendRequest(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	requests := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		http.Error(w, "unexpected request", http.StatusInternalServerError)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitem", "comment", "create", "--organization-id", "org-123", "--workitem-id", "wi-1", "--content", "hello", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), "yes")
	require.NotContains(t, stdout.String(), `"code": "AUTH_FAILED"`)
	require.Empty(t, stderr.String())
	require.Equal(t, 0, requests)
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
