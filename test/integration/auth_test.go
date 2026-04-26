package integration

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/stretchr/testify/require"
)

func TestAuthTopLevelNonInteractiveModePointsToTokenStdin(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()

	cmd := exec.Command(binary, "auth", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE=" + configDir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"code": "PARAM_INVALID"`)
	require.Contains(t, stdout.String(), "yunxiao auth login --token-stdin")
	require.NoFileExists(t, filepath.Join(configDir, "config.yaml"))
}

func TestAuthTopLevelPromptsForVisibleTokenAndWritesConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PTY integration test is Unix-only")
	}

	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()

	cmd := exec.Command(binary, "auth", "--skip-verify", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE=" + configDir)

	ptmx, err := pty.Start(cmd)
	require.NoError(t, err)
	defer ptmx.Close()

	var output bytes.Buffer
	require.Eventually(t, func() bool {
		buf := make([]byte, 128)
		_ = ptmx.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, _ := ptmx.Read(buf)
		if n > 0 {
			output.Write(buf[:n])
		}
		return strings.Contains(output.String(), "Enter Yunxiao access token:")
	}, 5*time.Second, 10*time.Millisecond)

	_, err = ptmx.Write([]byte("interactive-token\n"))
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		buf := make([]byte, 128)
		_ = ptmx.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, _ := ptmx.Read(buf)
		if n > 0 {
			output.Write(buf[:n])
		}
		return strings.Contains(output.String(), `"saved": true`)
	}, 5*time.Second, 10*time.Millisecond)

	require.Contains(t, output.String(), "interactive-token")

	err = cmd.Wait()
	require.NoError(t, err)

	config, err := os.ReadFile(filepath.Join(configDir, "config.yaml"))
	require.NoError(t, err)
	require.Contains(t, string(config), "access_token: interactive-token")
}

func TestAuthStatusReportsEnvSource(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)

	cmd := exec.Command(binary, "auth", "status", "--json")
	cmd.Env = testEnv("YUNXIAO_ACCESS_TOKEN=env-token")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"source":"env","configured":true,"verified":null},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
	require.NotContains(t, stdout.String(), "env-token")
}

func TestAuthStatusReportsConfigSource(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("access_token: config-token\n"), 0o600))

	cmd := exec.Command(binary, "auth", "status", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE=" + configDir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"source":"config","configured":true,"verified":null},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
	require.NotContains(t, stdout.String(), "config-token")
}

func TestAuthStatusReportsNoneSource(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()

	cmd := exec.Command(binary, "auth", "status", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE=" + configDir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"source":"none","configured":false,"verified":null},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestAuthStatusPrefersEnvOverConfig(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("access_token: config-token\n"), 0o600))

	cmd := exec.Command(binary, "auth", "status", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE="+configDir, "YUNXIAO_ACCESS_TOKEN=env-token")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"source":"env","configured":true,"verified":null},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
	require.NotContains(t, stdout.String(), "env-token")
	require.NotContains(t, stdout.String(), "config-token")
}

func TestAuthLoginWithoutTokenStdinFailsInNonInteractiveMode(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()

	cmd := exec.Command(binary, "auth", "login", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE=" + configDir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"code": "PARAM_INVALID"`)
	require.NoFileExists(t, filepath.Join(configDir, "config.yaml"))
}

func TestAuthLoginTokenStdinSkipVerifyWritesConfig(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()

	cmd := exec.Command(binary, "auth", "login", "--token-stdin", "--skip-verify", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE=" + configDir)
	cmd.Stdin = strings.NewReader("new-token\n")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"saved":true,"source":"config","verified":false,"path":"`+filepath.Join(configDir, "config.yaml")+`"},"meta":{},"error":null}`, stdout.String())
	require.NotContains(t, stdout.String(), "new-token")
	require.NotContains(t, stderr.String(), "new-token")

	config, err := os.ReadFile(filepath.Join(configDir, "config.yaml"))
	require.NoError(t, err)
	require.Contains(t, string(config), "access_token: new-token")

	configInfo, err := os.Stat(filepath.Join(configDir, "config.yaml"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), configInfo.Mode().Perm())

	dirInfo, err := os.Stat(configDir)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())
}

func TestAuthLoginTokenStdinRejectsEmbeddedNewline(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()

	cmd := exec.Command(binary, "auth", "login", "--token-stdin", "--skip-verify", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE=" + configDir)
	cmd.Stdin = strings.NewReader("line-one\nline-two\n")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"code": "PARAM_INVALID"`)
	require.NotContains(t, stdout.String(), "line-one")
	require.NotContains(t, stderr.String(), "line-one")
}

func TestAuthLogoutRemovesOnlyAccessToken(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("access_token: old-token\ntimeout: 9\norganization_id: org-1\n"), 0o600))

	cmd := exec.Command(binary, "auth", "logout", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE=" + configDir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"removed":true,"source":"config"},"meta":{},"error":null}`, stdout.String())
	require.NotContains(t, stdout.String(), "old-token")
	require.NotContains(t, stderr.String(), "old-token")

	config, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.NotContains(t, string(config), "access_token")
	require.Contains(t, string(config), "timeout: 9")
	require.Contains(t, string(config), "organization_id: org-1")
}

func TestAuthLogoutIsIdempotent(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("timeout: 9\n"), 0o600))

	cmd := exec.Command(binary, "auth", "logout", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE=" + configDir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"removed":false,"source":"config"},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestAuthLoginTokenStdinVerifiesBeforeWriting(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/oapi/v1/platform/user", r.URL.Path)
		require.Equal(t, "valid-token", r.Header.Get("x-yunxiao-token"))
		fmt.Fprint(w, `{"id":"user-1"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "auth", "login", "--token-stdin", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE="+configDir, "YUNXIAO_API_BASE_URL="+server.URL)
	cmd.Stdin = strings.NewReader("valid-token\n")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"saved":true,"source":"config","verified":true,"path":"`+filepath.Join(configDir, "config.yaml")+`"},"meta":{},"error":null}`, stdout.String())

	config, err := os.ReadFile(filepath.Join(configDir, "config.yaml"))
	require.NoError(t, err)
	require.Contains(t, string(config), "access_token: valid-token")
	require.NotContains(t, stdout.String(), "valid-token")
	require.NotContains(t, stderr.String(), "valid-token")
}

func TestAuthLoginVerificationFailureDoesNotWriteToken(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"message":"bad token"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "auth", "login", "--token-stdin", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE="+configDir, "YUNXIAO_API_BASE_URL="+server.URL)
	cmd.Stdin = strings.NewReader("bad-token\n")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 3, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"code": "AUTH_FAILED"`)
	require.NoFileExists(t, filepath.Join(configDir, "config.yaml"))
	require.NotContains(t, stdout.String(), "bad-token")
	require.NotContains(t, stderr.String(), "bad-token")
}

func TestAuthStatusVerifyUsesConfiguredToken(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("access_token: config-token\n"), 0o600))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/oapi/v1/platform/user", r.URL.Path)
		require.Equal(t, "config-token", r.Header.Get("x-yunxiao-token"))
		fmt.Fprint(w, `{"id":"user-1"}`)
	}))
	defer server.Close()

	cmd := exec.Command(binary, "auth", "status", "--verify", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE="+configDir, "YUNXIAO_API_BASE_URL="+server.URL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)
	require.JSONEq(t, `{"version":"v1","data":{"source":"config","configured":true,"verified":true},"meta":{},"error":null}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestAuthLoginRequiresForceToOverwriteConfigToken(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("access_token: old-token\n"), 0o600))

	cmd := exec.Command(binary, "auth", "login", "--token-stdin", "--skip-verify", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE=" + configDir)
	cmd.Stdin = strings.NewReader("new-token\n")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok)
	require.Equal(t, 2, exitErr.ExitCode())
	require.Contains(t, stdout.String(), `"code": "PARAM_INVALID"`)

	config, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.Contains(t, string(config), "access_token: old-token")
	require.NotContains(t, string(config), "new-token")
}

func TestAuthLoginForceOverwritesConfigToken(t *testing.T) {
	root := filepath.Join("..", "..")
	binary := buildTestBinary(t, root)
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("access_token: old-token\ntimeout: 9\n"), 0o600))

	cmd := exec.Command(binary, "auth", "login", "--token-stdin", "--skip-verify", "--force", "--json")
	cmd.Env = testEnv("YUNXIAO_CONFIG_FILE=" + configDir)
	cmd.Stdin = strings.NewReader("new-token\n")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err)

	config, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.Contains(t, string(config), "access_token: new-token")
	require.Contains(t, string(config), "timeout: 9")
	require.NotContains(t, string(config), "old-token")
}
