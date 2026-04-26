package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodeupBranchesListCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/codeup/repositories/repo-1/branches", r.URL.Path)
		require.Equal(t, "valid-token", r.Header.Get("x-yunxiao-token"))
		require.Equal(t, "50", r.URL.Query().Get("perPage"))
		require.Equal(t, "2", r.URL.Query().Get("page"))
		require.Equal(t, "name", r.URL.Query().Get("sort"))
		require.Equal(t, "main", r.URL.Query().Get("search"))
		w.Header().Set("x-next-page", "3")
		w.Header().Set("x-per-page", "50")
		fmt.Fprint(w, `[{"name":"main","protected":true}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "codeup", "branches", "list", "--organization-id", "org-123", "--repo-id", "repo-1", "--page-size", "50", "--page-token", "2", "--sort", "name", "--search", "main", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	require.NoError(t, cmd.Run())
	require.JSONEq(t, `{"version":"v1","data":[{"name":"main","protected":true}],"meta":{"pagination":{"next_token":"3","page_size":50,"has_more":true}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestCodeupCommitsListCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/codeup/repositories/repo-1/commits", r.URL.Path)
		query := r.URL.Query()
		require.Equal(t, "20", query.Get("perPage"))
		require.Equal(t, "4", query.Get("page"))
		require.Equal(t, "main", query.Get("refName"))
		require.Equal(t, "2026-04-01", query.Get("since"))
		require.Equal(t, "2026-04-26", query.Get("until"))
		require.Equal(t, "cmd/yunxiao/main.go", query.Get("path"))
		require.Equal(t, "fix", query.Get("search"))
		require.Equal(t, "true", query.Get("showSignature"))
		require.Equal(t, "u1,u2", query.Get("committerIds"))
		w.Header().Set("x-next-page", "5")
		fmt.Fprint(w, `[{"id":"c1","shortId":"abc123","message":"fix bug"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "codeup", "commits", "list", "--organization-id", "org-123", "--repo-id", "repo-1", "--page-size", "20", "--page-token", "4", "--ref-name", "main", "--since", "2026-04-01", "--until", "2026-04-26", "--path", "cmd/yunxiao/main.go", "--search", "fix", "--show-signature", "--committer-ids", "u1,u2", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	require.NoError(t, cmd.Run())
	require.JSONEq(t, `{"version":"v1","data":[{"id":"c1","shortId":"abc123","message":"fix bug"}],"meta":{"pagination":{"next_token":"5","page_size":20,"has_more":true}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestCodeupFileGetCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/codeup/repositories/repo-1/files/README.md", r.URL.Path)
		require.Equal(t, "main", r.URL.Query().Get("ref"))
		fmt.Fprint(w, `{"fileName":"README.md","content":"IyBkZW1v"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "codeup", "file", "get", "--organization-id", "org-123", "--repo-id", "repo-1", "--path", "README.md", "--ref", "main", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	require.NoError(t, cmd.Run())
	require.JSONEq(t, `{"version":"v1","data":{"fileName":"README.md","content":"IyBkZW1v"},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestCodeupCompareGetCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/codeup/repositories/repo-1/compares", r.URL.Path)
		query := r.URL.Query()
		require.Equal(t, "main", query.Get("from"))
		require.Equal(t, "feature", query.Get("to"))
		require.Equal(t, "branch", query.Get("sourceType"))
		require.Equal(t, "branch", query.Get("targetType"))
		require.Equal(t, "true", query.Get("straight"))
		fmt.Fprint(w, `{"commitCount":2,"commits":[{"id":"c1"}]}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "codeup", "compare", "get", "--organization-id", "org-123", "--repo-id", "repo-1", "--from", "main", "--to", "feature", "--source-type", "branch", "--target-type", "branch", "--straight", "true", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	require.NoError(t, cmd.Run())
	require.JSONEq(t, `{"version":"v1","data":{"commitCount":2,"commits":[{"id":"c1"}]},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestFlowRunsListCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/flow/pipelines/p-1/runs", r.URL.Path)
		query := r.URL.Query()
		require.Equal(t, "10", query.Get("perPage"))
		require.Equal(t, "2", query.Get("page"))
		require.Equal(t, "2026-04-01", query.Get("startTime"))
		require.Equal(t, "2026-04-26", query.Get("endTime"))
		require.Equal(t, "SUCCESS", query.Get("status"))
		require.Equal(t, "MANUAL", query.Get("triggerMode"))
		w.Header().Set("x-next-page", "3")
		fmt.Fprint(w, `[{"id":"run-1","status":"SUCCESS"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "flow", "runs", "list", "--organization-id", "org-123", "--pipeline-id", "p-1", "--page-size", "10", "--page-token", "2", "--start-time", "2026-04-01", "--end-time", "2026-04-26", "--status", "SUCCESS", "--trigger-mode", "MANUAL", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	require.NoError(t, cmd.Run())
	require.JSONEq(t, `{"version":"v1","data":[{"id":"run-1","status":"SUCCESS"}],"meta":{"pagination":{"next_token":"3","page_size":10,"has_more":true}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestFlowRunGetCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/flow/pipelines/p-1/runs/run-1", r.URL.Path)
		fmt.Fprint(w, `{"id":"run-1","status":"SUCCESS"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "flow", "run", "get", "--organization-id", "org-123", "--pipeline-id", "p-1", "--run-id", "run-1", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	require.NoError(t, cmd.Run())
	require.JSONEq(t, `{"version":"v1","data":{"id":"run-1","status":"SUCCESS"},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestOrgMembersListCallsYunxiaoAPI(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/oapi/v1/platform/members", r.URL.Path)
		require.Equal(t, "100", r.URL.Query().Get("perPage"))
		require.Equal(t, "2", r.URL.Query().Get("page"))
		w.Header().Set("x-next-page", "3")
		fmt.Fprint(w, `[{"id":"m-1","name":"Nick"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "org", "members", "list", "--organization-id", "org-123", "--page-size", "100", "--page-token", "2", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	require.NoError(t, cmd.Run())
	require.JSONEq(t, `{"version":"v1","data":[{"id":"m-1","name":"Nick"}],"meta":{"pagination":{"next_token":"3","page_size":100,"has_more":true}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestProjexWorkitemsListSendsMCPCompatibleFilters(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/oapi/v1/projex/workitems:search", r.URL.Path)

		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "Task", payload["category"])
		require.Equal(t, "proj-1", payload["spaceId"])
		require.Equal(t, float64(30), payload["perPage"])
		require.Equal(t, "finishTime", payload["orderBy"])
		require.Equal(t, "desc", payload["sort"])

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
		fields := map[string]struct {
			Operator string
			ToValue  *string
			Value    []string
		}{}
		for _, condition := range conditions.ConditionGroups[0] {
			fields[condition.FieldIdentifier] = struct {
				Operator string
				ToValue  *string
				Value    []string
			}{Operator: condition.Operator, ToValue: condition.ToValue, Value: condition.Value}
		}
		require.Equal(t, "CONTAINS", fields["subject"].Operator)
		require.Equal(t, []string{"bug"}, fields["subject"].Value)
		require.Equal(t, "CONTAINS", fields["status"].Operator)
		require.Equal(t, []string{"s1", "s2"}, fields["status"].Value)
		require.Equal(t, "CONTAINS", fields["assignedTo"].Operator)
		require.Equal(t, []string{"u1", "u2"}, fields["assignedTo"].Value)
		require.Equal(t, "BETWEEN", fields["finishTime"].Operator)
		require.Equal(t, []string{"2026-04-20 00:00:00"}, fields["finishTime"].Value)
		require.Equal(t, "BETWEEN", fields["updateStatusAt"].Operator)
		require.Nil(t, fields["updateStatusAt"].Value)
		require.NotNil(t, fields["updateStatusAt"].ToValue)
		require.Equal(t, "2026-04-25 23:59:59", *fields["updateStatusAt"].ToValue)

		fmt.Fprint(w, `[{"id":"wi-1","subject":"fix bug"}]`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "projex", "workitems", "list", "--organization-id", "org-123", "--category", "Task", "--space-id", "proj-1", "--page-size", "30", "--subject", "bug", "--status", "s1,s2", "--assigned-to", "u1,u2", "--finish-time-after", "2026-04-20", "--finish-time-before", "2026-04-26", "--update-status-at-before", "2026-04-25", "--order-by", "finishTime", "--sort", "desc", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=valid-token", "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	require.NoError(t, cmd.Run())
	require.JSONEq(t, `{"version":"v1","data":[{"id":"wi-1","subject":"fix bug"}],"meta":{"pagination":{"next_token":null,"page_size":30,"has_more":false}},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}
