package integration

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnknownFlagReturnsJSONEnvelope(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "org", "current", "--json", "--bad-flag")
	cmd.Env = testEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"version": "v1"`)
	require.Contains(t, stdout.String(), `"category": "param"`)
	require.Contains(t, stderr.String(), "unknown flag")
}

func TestOrgMembersListParameterErrorIncludesTraceID(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "org", "members", "list", "--json", "--trace-id", "trace-param")
	cmd.Env = testEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"trace_id": "trace-param"`)
	require.Contains(t, stdout.String(), `"code": "PARAM_REQUIRED"`)
	require.Empty(t, stderr.String())
}

func TestUnknownCommandReturnsJSONEnvelope(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "unknown", "--json")
	cmd.Env = testEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"version": "v1"`)
	require.Contains(t, stdout.String(), `"category": "param"`)
	require.Contains(t, stdout.String(), `"code": "COMMAND_FAILED"`)
	require.Contains(t, stderr.String(), "unknown command")
}

func TestCommandsJSONIncludesPhase1Commands(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "commands", "--json")
	cmd.Env = testEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.Empty(t, stderr.String())
	require.Contains(t, stdout.String(), `"path": "yunxiao org current"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao org members list"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao codeup repos list"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao codeup repo get"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao codeup branches list"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao codeup commits list"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao codeup file get"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao codeup compare get"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao flow pipelines list"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao flow pipeline get"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao flow runs list"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao flow run get"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao projex workitem create"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao projex workitem update"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao projex workitem-types list"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao projex workitem-type get"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao projex workitem-types relations"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao projex workitem-type fields"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao projex workitem-type workflow"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao projex project-templates list"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao projex project-template fields"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao projex project create"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao projex project archive"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao projex workitem comments list"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao projex workitem comment create"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao auth login"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao auth status"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao auth logout"`)
	require.Contains(t, stdout.String(), `"flags"`)
	require.Contains(t, stdout.String(), `"name": "organization-id"`)
	require.Contains(t, stdout.String(), `"required": true`)
	require.Contains(t, stdout.String(), `"type": "string"`)
}

func TestCommandHelpJSONIncludesFlags(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "flow", "pipeline", "get", "--help", "--json")
	cmd.Env = testEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.Empty(t, stderr.String())
	require.Contains(t, stdout.String(), `"version": "v1"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao flow pipeline get"`)
	require.Contains(t, stdout.String(), `"name": "pipeline-id"`)
	require.Contains(t, stdout.String(), `"required": true`)
	require.Contains(t, stdout.String(), `"type": "string"`)
}

func TestPhase1GapCommandHelpJSONIncludesRequiredFlags(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	tests := []struct {
		name     string
		args     []string
		required []string
		optional []string
	}{
		{name: "org members list", args: []string{"org", "members", "list"}, required: []string{"organization-id"}, optional: []string{"page-size", "page-token"}},
		{name: "codeup branches list", args: []string{"codeup", "branches", "list"}, required: []string{"organization-id", "repo-id"}, optional: []string{"page-size", "page-token", "sort", "search"}},
		{name: "codeup commits list", args: []string{"codeup", "commits", "list"}, required: []string{"organization-id", "repo-id"}, optional: []string{"ref-name", "since", "until", "path", "search", "show-signature", "committer-ids"}},
		{name: "codeup file get", args: []string{"codeup", "file", "get"}, required: []string{"organization-id", "repo-id", "path", "ref"}},
		{name: "codeup compare get", args: []string{"codeup", "compare", "get"}, required: []string{"organization-id", "repo-id", "from", "to"}, optional: []string{"source-type", "target-type", "straight"}},
		{name: "flow runs list", args: []string{"flow", "runs", "list"}, required: []string{"organization-id", "pipeline-id"}, optional: []string{"page-size", "page-token", "start-time", "end-time", "status", "trigger-mode"}},
		{name: "flow run get", args: []string{"flow", "run", "get"}, required: []string{"organization-id", "pipeline-id", "run-id"}},
		{name: "projex projects list", args: []string{"projex", "projects", "list"}, optional: []string{"organization-id", "mine", "scenario-filter", "user-id", "name", "status", "created-after", "created-before", "admin-user-id", "advanced-conditions", "extra-conditions", "order-by", "sort"}},
		{name: "projex workitems list", args: []string{"projex", "workitems", "list"}, required: []string{"category"}, optional: []string{"organization-id", "project-id", "space-id", "mine", "unfinished", "status", "assigned-to", "finish-time-after", "update-status-at-after", "advanced-conditions", "order-by", "sort"}},
		{name: "projex workitem create", args: []string{"projex", "workitem", "create"}, required: []string{"organization-id", "workitem-type-id", "subject", "assigned-to", "yes"}, optional: []string{"project-id", "space-id", "description", "description-file", "format-type", "labels", "parent-id", "participants", "sprint", "trackers", "verifier", "versions", "custom-field", "custom-fields-json"}},
		{name: "projex workitem update", args: []string{"projex", "workitem", "update"}, required: []string{"organization-id", "workitem-id", "yes"}, optional: []string{"subject", "assigned-to", "description", "description-file", "format-type", "priority", "status", "labels", "participants", "sprint", "trackers", "verifier", "versions", "custom-field", "custom-fields-json"}},
		{name: "projex project-templates list", args: []string{"projex", "project-templates", "list"}, required: []string{"organization-id"}},
		{name: "projex project-template fields", args: []string{"projex", "project-template", "fields"}, required: []string{"organization-id", "template-id"}},
		{name: "projex project create", args: []string{"projex", "project", "create"}, required: []string{"organization-id", "name", "custom-code", "template-id", "scope", "yes"}, optional: []string{"description", "description-file", "custom-field", "custom-fields-json"}},
		{name: "projex project archive", args: []string{"projex", "project", "archive"}, required: []string{"organization-id", "project-id", "yes"}, optional: []string{"operator-id"}},
		{name: "projex workitem-types list all", args: []string{"projex", "workitem-types", "list"}, required: []string{"organization-id"}, optional: []string{"all", "project-id", "category"}},
		{name: "projex workitem-type get", args: []string{"projex", "workitem-type", "get"}, required: []string{"organization-id", "workitem-type-id"}},
		{name: "projex workitem-types relations", args: []string{"projex", "workitem-types", "relations"}, required: []string{"organization-id", "workitem-type-id"}},
		{name: "projex workitem-type fields", args: []string{"projex", "workitem-type", "fields"}, required: []string{"organization-id", "project-id", "workitem-type-id"}},
		{name: "projex workitem-type workflow", args: []string{"projex", "workitem-type", "workflow"}, required: []string{"organization-id", "project-id", "workitem-type-id"}},
		{name: "projex workitem comments list", args: []string{"projex", "workitem", "comments", "list"}, required: []string{"organization-id", "workitem-id"}, optional: []string{"page-size", "page-token"}},
		{name: "projex workitem comment create", args: []string{"projex", "workitem", "comment", "create"}, required: []string{"organization-id", "workitem-id", "yes"}, optional: []string{"content", "content-file"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := append(append([]string{}, tc.args...), "--help", "--json")
			cmd := exec.Command(binary, args...)
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
						Name     string `json:"name"`
						Required bool   `json:"required"`
					} `json:"flags"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(stdout.Bytes(), &envelope))

			flags := make(map[string]bool, len(envelope.Data.Flags))
			for _, flag := range envelope.Data.Flags {
				flags[flag.Name] = flag.Required
			}
			for _, name := range tc.required {
				required, ok := flags[name]
				require.True(t, ok, "expected flag %s", name)
				require.True(t, required, "expected flag %s to be required", name)
			}
			for _, name := range tc.optional {
				_, ok := flags[name]
				require.True(t, ok, "expected flag %s", name)
			}
		})
	}
}

func TestAuthHelpJSONIncludesAuthFlags(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	t.Run("auth", func(t *testing.T) {
		authCmd := exec.Command(binary, "auth", "--help", "--json")
		authCmd.Env = testEnv()

		var stdout, stderr bytes.Buffer
		authCmd.Stdout = &stdout
		authCmd.Stderr = &stderr

		err := authCmd.Run()
		require.NoError(t, err)
		require.Empty(t, stderr.String())
		require.Contains(t, stdout.String(), `"path": "yunxiao auth"`)
		require.Contains(t, stdout.String(), `"name": "skip-verify"`)
		require.Contains(t, stdout.String(), `"name": "force"`)
	})

	t.Run("login", func(t *testing.T) {
		loginCmd := exec.Command(binary, "auth", "login", "--help", "--json")
		loginCmd.Env = testEnv()

		var stdout, stderr bytes.Buffer
		loginCmd.Stdout = &stdout
		loginCmd.Stderr = &stderr

		err := loginCmd.Run()
		require.NoError(t, err)
		require.Empty(t, stderr.String())
		require.Contains(t, stdout.String(), `"path": "yunxiao auth login"`)
		require.Contains(t, stdout.String(), `"name": "token-stdin"`)
		require.Contains(t, stdout.String(), `"name": "skip-verify"`)
		require.Contains(t, stdout.String(), `"name": "force"`)
	})

	t.Run("status", func(t *testing.T) {
		statusCmd := exec.Command(binary, "auth", "status", "--help", "--json")
		statusCmd.Env = testEnv()

		var stdout, stderr bytes.Buffer
		statusCmd.Stdout = &stdout
		statusCmd.Stderr = &stderr

		err := statusCmd.Run()
		require.NoError(t, err)
		require.Empty(t, stderr.String())
		require.Contains(t, stdout.String(), `"path": "yunxiao auth status"`)
		require.Contains(t, stdout.String(), `"name": "verify"`)
	})
}

func TestTestplansHelpDoesNotExposePageSize(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "testhub", "testplans", "list", "--help", "--json")
	cmd.Env = testEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.Empty(t, stderr.String())

	var envelope struct {
		Data struct {
			Path  string `json:"path"`
			Flags []struct {
				Name string `json:"name"`
			} `json:"flags"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &envelope))
	require.Equal(t, "yunxiao testhub testplans list", envelope.Data.Path)
	for _, flag := range envelope.Data.Flags {
		require.NotEqual(t, "page-size", flag.Name)
	}
}
