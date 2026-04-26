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
	require.Contains(t, stdout.String(), `"path": "yunxiao codeup repos list"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao codeup repo get"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao flow pipelines list"`)
	require.Contains(t, stdout.String(), `"path": "yunxiao flow pipeline get"`)
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

func TestAuthHelpJSONIncludesAuthFlags(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	authCmd := exec.Command(binary, "auth", "--help", "--json")
	authCmd.Env = testEnv()

	var authStdout, authStderr bytes.Buffer
	authCmd.Stdout = &authStdout
	authCmd.Stderr = &authStderr

	err := authCmd.Run()
	require.NoError(t, err)
	require.Empty(t, authStderr.String())
	require.Contains(t, authStdout.String(), `"path": "yunxiao auth"`)
	require.Contains(t, authStdout.String(), `"name": "skip-verify"`)
	require.Contains(t, authStdout.String(), `"name": "force"`)

	loginCmd := exec.Command(binary, "auth", "login", "--help", "--json")
	loginCmd.Env = testEnv()

	var loginStdout, loginStderr bytes.Buffer
	loginCmd.Stdout = &loginStdout
	loginCmd.Stderr = &loginStderr

	err = loginCmd.Run()
	require.NoError(t, err)
	require.Empty(t, loginStderr.String())
	require.Contains(t, loginStdout.String(), `"path": "yunxiao auth login"`)
	require.Contains(t, loginStdout.String(), `"name": "token-stdin"`)
	require.Contains(t, loginStdout.String(), `"name": "skip-verify"`)
	require.Contains(t, loginStdout.String(), `"name": "force"`)

	statusCmd := exec.Command(binary, "auth", "status", "--help", "--json")
	statusCmd.Env = testEnv()

	var statusStdout, statusStderr bytes.Buffer
	statusCmd.Stdout = &statusStdout
	statusCmd.Stderr = &statusStderr

	err = statusCmd.Run()
	require.NoError(t, err)
	require.Empty(t, statusStderr.String())
	require.Contains(t, statusStdout.String(), `"path": "yunxiao auth status"`)
	require.Contains(t, statusStdout.String(), `"name": "verify"`)
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
