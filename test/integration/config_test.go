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
	"time"

	"github.com/stretchr/testify/require"
)

func TestMalformedConfigReturnsJSONEnvelope(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	tmpDir := t.TempDir()
	badConfigDir := filepath.Join(tmpDir, "config")
	require.NoError(t, os.MkdirAll(badConfigDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(badConfigDir, "config.yaml"), []byte("api_base_url: ["), 0o644))

	cmd := exec.Command(binary, "commands", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE=" + badConfigDir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 1, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"code": "CONFIG_READ_FAILED"`)
	require.Contains(t, stdout.String(), `"category": "general"`)
	require.Contains(t, stderr.String(), "failed to read config")
}

func TestConfigTimeoutAppliesWhenFlagIsOmitted(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("timeout: 1\n"), 0o644))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, `{}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "org", "current", "--json")
	cmd.Env = testEnv(
		"YUNXIAO_CONFIG_FILE="+configDir,
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
	require.Contains(t, stdout.String(), `"code": "REQUEST_TIMEOUT"`)
	require.Contains(t, stdout.String(), "after 1s")
	require.Contains(t, stderr.String(), "org current failed")
}

func TestRegionDefaultOrganizationIDAppliesWhenFlagIsOmitted(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/oapi/v1/codeup/repositories", r.URL.Path)
		fmt.Fprint(w, `{"data":[],"nextPage":null}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "codeup", "repos", "list", "--json")
	cmd.Env = testEnv(
		"YUNXIAO_ACCESS_TOKEN=valid-token",
		"YUNXIAO_API_BASE_URL="+server.URL,
		"YUNXIAO_REGION_DEFAULT_ORG_ID=org-from-region",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.Contains(t, stdout.String(), `"data": []`)
	require.Empty(t, stderr.String())
}
