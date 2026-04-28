package installer_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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

	"github.com/stretchr/testify/require"
)

type releaseFixture struct {
	asset    string
	archive  []byte
	checksum string
}

func TestPOSIXInstallUpdateUninstallWithFixtures(t *testing.T) {
	root := repoRoot(t)
	server := fixtureServer(t, "v1.2.3", map[string]releaseFixture{
		"v1.2.3": newTarRelease(t, "v1.2.3", fixtureShellBinary("fixture-v1")),
		"v1.2.4": newTarRelease(t, "v1.2.4", fixtureShellBinary("fixture-v2")),
	})

	binDir := filepath.Join(t.TempDir(), "bin")
	env := installerEnv(t, map[string]string{
		"YUNXIAO_INSTALL_OS":                "linux",
		"YUNXIAO_INSTALL_ARCH":              "amd64",
		"YUNXIAO_INSTALL_API_URL":           server.URL,
		"YUNXIAO_INSTALL_DOWNLOAD_BASE_URL": server.URL,
	})

	stdout, stderr, code := runCommand(t, env, "sh", filepath.Join(root, "scripts", "install.sh"), "--bin-dir", binDir)
	require.Equal(t, 0, code, "stdout=%s stderr=%s", stdout, stderr)
	target := filepath.Join(binDir, "yunxiao")
	require.FileExists(t, target)
	require.Contains(t, readFile(t, target), "fixture-v1")

	stdout, stderr, code = runCommand(t, env, "sh", filepath.Join(root, "scripts", "install.sh"), "--version", "v1.2.4", "--bin-dir", binDir)
	require.Equal(t, 0, code, "stdout=%s stderr=%s", stdout, stderr)
	require.Contains(t, readFile(t, target), "fixture-v2")

	stdout, stderr, code = runCommand(t, env, "sh", filepath.Join(root, "scripts", "install.sh"), "--bin-dir", binDir, "--uninstall")
	require.Equal(t, 0, code, "stdout=%s stderr=%s", stdout, stderr)
	require.NoFileExists(t, target)

	stdout, stderr, code = runCommand(t, env, "sh", filepath.Join(root, "scripts", "install.sh"), "--bin-dir", binDir, "--uninstall")
	require.Equal(t, 0, code, "stdout=%s stderr=%s", stdout, stderr)
	require.Contains(t, stderr, "already not installed")
}

func TestPOSIXChecksumMismatchDoesNotOverwrite(t *testing.T) {
	root := repoRoot(t)
	fixture := newTarRelease(t, "v1.2.3", fixtureShellBinary("new-content"))
	fixture.checksum = strings.Repeat("0", 64)
	server := fixtureServer(t, "v1.2.3", map[string]releaseFixture{"v1.2.3": fixture})
	binDir := filepath.Join(t.TempDir(), "bin")
	target := filepath.Join(binDir, "yunxiao")
	require.NoError(t, os.MkdirAll(binDir, 0o755))
	require.NoError(t, os.WriteFile(target, []byte("old-content"), 0o755))

	env := installerEnv(t, map[string]string{
		"YUNXIAO_INSTALL_OS":                "linux",
		"YUNXIAO_INSTALL_ARCH":              "amd64",
		"YUNXIAO_INSTALL_DOWNLOAD_BASE_URL": server.URL,
	})

	stdout, stderr, code := runCommand(t, env, "sh", filepath.Join(root, "scripts", "install.sh"), "--version", "v1.2.3", "--bin-dir", binDir)
	require.Equal(t, 4, code, "stdout=%s stderr=%s", stdout, stderr)
	require.Contains(t, stderr, "checksum mismatch")
	require.Equal(t, "old-content", readFile(t, target))
}

func TestPOSIXArchiveMissingBinaryFailsBeforeInstall(t *testing.T) {
	root := repoRoot(t)
	server := fixtureServer(t, "v1.2.3", map[string]releaseFixture{
		"v1.2.3": newTarReleaseWithEntries(t, "v1.2.3", map[string]string{"not-yunxiao": "wrong"}),
	})
	binDir := filepath.Join(t.TempDir(), "bin")

	env := installerEnv(t, map[string]string{
		"YUNXIAO_INSTALL_OS":                "linux",
		"YUNXIAO_INSTALL_ARCH":              "amd64",
		"YUNXIAO_INSTALL_DOWNLOAD_BASE_URL": server.URL,
	})

	stdout, stderr, code := runCommand(t, env, "sh", filepath.Join(root, "scripts", "install.sh"), "--version", "v1.2.3", "--bin-dir", binDir)
	require.Equal(t, 1, code, "stdout=%s stderr=%s", stdout, stderr)
	require.Contains(t, stderr, "archive does not contain expected yunxiao binary")
	require.NoFileExists(t, filepath.Join(binDir, "yunxiao"))
}

func TestPOSIXVerificationFailureDoesNotOverwrite(t *testing.T) {
	root := repoRoot(t)
	server := fixtureServer(t, "v1.2.3", map[string]releaseFixture{
		"v1.2.3": newTarRelease(t, "v1.2.3", fixtureFailingVerifyBinary("bad-content")),
	})
	binDir := filepath.Join(t.TempDir(), "bin")
	target := filepath.Join(binDir, "yunxiao")
	require.NoError(t, os.MkdirAll(binDir, 0o755))
	require.NoError(t, os.WriteFile(target, []byte("old-content"), 0o755))

	env := installerEnv(t, map[string]string{
		"YUNXIAO_INSTALL_OS":                "linux",
		"YUNXIAO_INSTALL_ARCH":              "amd64",
		"YUNXIAO_INSTALL_DOWNLOAD_BASE_URL": server.URL,
	})

	stdout, stderr, code := runCommand(t, env, "sh", filepath.Join(root, "scripts", "install.sh"), "--version", "v1.2.3", "--bin-dir", binDir)
	require.Equal(t, 6, code, "stdout=%s stderr=%s", stdout, stderr)
	require.Contains(t, stderr, "failed verification before install")
	require.Equal(t, "old-content", readFile(t, target))
}

func TestPOSIXDryRunMappingAndDirectoryPriority(t *testing.T) {
	root := repoRoot(t)
	home := t.TempDir()
	envDir := filepath.Join(t.TempDir(), "env-bin")
	flagDir := filepath.Join(t.TempDir(), "flag-bin")
	env := installerEnv(t, map[string]string{
		"HOME":                              home,
		"YUNXIAO_INSTALL_OS":                "linux",
		"YUNXIAO_INSTALL_ARCH":              "arm64",
		"YUNXIAO_INSTALL_REPO":              "mirror/yunxiao-cli",
		"YUNXIAO_INSTALL_DIR":               envDir,
		"YUNXIAO_INSTALL_DOWNLOAD_BASE_URL": "https://downloads.example.test",
	})

	stdout, stderr, code := runCommand(t, env, "sh", filepath.Join(root, "scripts", "install.sh"), "--version", "v9.9.9", "--bin-dir", flagDir, "--dry-run")
	require.Equal(t, 0, code, "stdout=%s stderr=%s", stdout, stderr)
	require.Contains(t, stdout, "os=linux")
	require.Contains(t, stdout, "arch=arm64")
	require.Contains(t, stdout, "asset=yunxiao_v9.9.9_linux_arm64.tar.gz")
	require.Contains(t, stdout, "archive_url=https://downloads.example.test/mirror/yunxiao-cli/releases/download/v9.9.9/yunxiao_v9.9.9_linux_arm64.tar.gz")
	require.Contains(t, normalizeSlashes(stdout), "target="+filepath.ToSlash(filepath.Join(flagDir, "yunxiao")))

	stdout, stderr, code = runCommand(t, env, "sh", filepath.Join(root, "scripts", "install.sh"), "--version", "v9.9.9", "--dry-run")
	require.Equal(t, 0, code, "stdout=%s stderr=%s", stdout, stderr)
	require.Contains(t, normalizeSlashes(stdout), "target="+filepath.ToSlash(filepath.Join(envDir, "yunxiao")))
}

func TestPOSIXUnsupportedAndUnsafeTargets(t *testing.T) {
	root := repoRoot(t)
	env := installerEnv(t, map[string]string{
		"YUNXIAO_INSTALL_OS":   "freebsd",
		"YUNXIAO_INSTALL_ARCH": "amd64",
	})
	stdout, stderr, code := runCommand(t, env, "sh", filepath.Join(root, "scripts", "install.sh"), "--version", "v1.2.3", "--dry-run")
	require.Equal(t, 2, code, "stdout=%s stderr=%s", stdout, stderr)
	require.NotContains(t, stdout, "archive_url=")

	env = installerEnv(t, map[string]string{
		"YUNXIAO_INSTALL_OS":   "linux",
		"YUNXIAO_INSTALL_ARCH": "amd64",
	})
	stdout, stderr, code = runCommand(t, env, "sh", filepath.Join(root, "scripts", "install.sh"), "--bin-dir", "/", "--uninstall")
	require.Equal(t, 7, code, "stdout=%s stderr=%s", stdout, stderr)
	require.Contains(t, stderr, "refusing to uninstall from root directory")
}

func TestPOSIXUninstallRefusesSymlinkDirectoryWhenAvailable(t *testing.T) {
	root := repoRoot(t)
	realDir := t.TempDir()
	linkDir := filepath.Join(t.TempDir(), "bin-link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("directory symlink is not available: %v", err)
	}
	env := installerEnv(t, map[string]string{
		"YUNXIAO_INSTALL_OS":   "linux",
		"YUNXIAO_INSTALL_ARCH": "amd64",
	})

	stdout, stderr, code := runCommand(t, env, "sh", filepath.Join(root, "scripts", "install.sh"), "--bin-dir", linkDir, "--uninstall")
	require.Equal(t, 7, code, "stdout=%s stderr=%s", stdout, stderr)
	require.Contains(t, stderr, "refusing to uninstall from symlink directory")
}

func TestPowerShellInstallUpdateUninstallWithZipFixtureWhenAvailable(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell install fixture requires Windows to execute yunxiao.exe")
	}
	pwsh := requirePowerShell(t)
	root := repoRoot(t)
	server := fixtureServer(t, "v1.2.3", map[string]releaseFixture{
		"v1.2.3": newZipRelease(t, "v1.2.3", windowsFixtureBinaryWithMarker(t, "ps-fixture-v1")),
		"v1.2.4": newZipRelease(t, "v1.2.4", windowsFixtureBinaryWithMarker(t, "ps-fixture-v2")),
	})
	binDir := filepath.Join(t.TempDir(), "bin")
	env := installerEnv(t, map[string]string{
		"USERPROFILE":                       t.TempDir(),
		"YUNXIAO_INSTALL_OS":                "windows",
		"YUNXIAO_INSTALL_ARCH":              "amd64",
		"YUNXIAO_INSTALL_API_URL":           server.URL,
		"YUNXIAO_INSTALL_DOWNLOAD_BASE_URL": server.URL,
	})

	stdout, stderr, code := runCommand(t, env, pwsh, "-NoProfile", "-File", filepath.Join(root, "scripts", "install.ps1"), "-BinDir", binDir)
	require.Equal(t, 0, code, "stdout=%s stderr=%s", stdout, stderr)
	target := filepath.Join(binDir, "yunxiao.exe")
	require.FileExists(t, target)
	require.Contains(t, readFile(t, target), "ps-fixture-v1")

	stdout, stderr, code = runCommand(t, env, pwsh, "-NoProfile", "-File", filepath.Join(root, "scripts", "install.ps1"), "-Version", "v1.2.4", "-BinDir", binDir)
	require.Equal(t, 0, code, "stdout=%s stderr=%s", stdout, stderr)
	require.Contains(t, readFile(t, target), "ps-fixture-v2")

	stdout, stderr, code = runCommand(t, env, pwsh, "-NoProfile", "-File", filepath.Join(root, "scripts", "install.ps1"), "-BinDir", binDir, "-Uninstall")
	require.Equal(t, 0, code, "stdout=%s stderr=%s", stdout, stderr)
	require.NoFileExists(t, target)
}

func TestPowerShellChecksumMismatchDoesNotOverwriteWhenAvailable(t *testing.T) {
	pwsh := requirePowerShell(t)
	root := repoRoot(t)
	fixture := newZipRelease(t, "v1.2.3", []byte("new-content"))
	fixture.checksum = strings.Repeat("0", 64)
	server := fixtureServer(t, "v1.2.3", map[string]releaseFixture{"v1.2.3": fixture})
	binDir := filepath.Join(t.TempDir(), "bin")
	target := filepath.Join(binDir, "yunxiao.exe")
	require.NoError(t, os.MkdirAll(binDir, 0o755))
	require.NoError(t, os.WriteFile(target, []byte("old-content"), 0o755))
	env := installerEnv(t, map[string]string{
		"USERPROFILE":                       t.TempDir(),
		"YUNXIAO_INSTALL_OS":                "windows",
		"YUNXIAO_INSTALL_ARCH":              "amd64",
		"YUNXIAO_INSTALL_DOWNLOAD_BASE_URL": server.URL,
	})

	stdout, stderr, code := runCommand(t, env, pwsh, "-NoProfile", "-File", filepath.Join(root, "scripts", "install.ps1"), "-Version", "v1.2.3", "-BinDir", binDir)
	require.Equal(t, 4, code, "stdout=%s stderr=%s", stdout, stderr)
	require.Contains(t, stderr, "checksum mismatch")
	require.Equal(t, "old-content", readFile(t, target))
}

func TestPowerShellArchiveMissingBinaryFailsBeforeInstallWhenAvailable(t *testing.T) {
	pwsh := requirePowerShell(t)
	root := repoRoot(t)
	server := fixtureServer(t, "v1.2.3", map[string]releaseFixture{
		"v1.2.3": newZipReleaseWithEntries(t, "v1.2.3", map[string][]byte{"not-yunxiao.exe": []byte("wrong")}),
	})
	binDir := filepath.Join(t.TempDir(), "bin")
	env := installerEnv(t, map[string]string{
		"USERPROFILE":                       t.TempDir(),
		"YUNXIAO_INSTALL_OS":                "windows",
		"YUNXIAO_INSTALL_ARCH":              "amd64",
		"YUNXIAO_INSTALL_DOWNLOAD_BASE_URL": server.URL,
	})

	stdout, stderr, code := runCommand(t, env, pwsh, "-NoProfile", "-File", filepath.Join(root, "scripts", "install.ps1"), "-Version", "v1.2.3", "-BinDir", binDir)
	require.Equal(t, 1, code, "stdout=%s stderr=%s", stdout, stderr)
	require.Contains(t, stderr, "archive does not contain expected yunxiao.exe binary")
	require.NoFileExists(t, filepath.Join(binDir, "yunxiao.exe"))
}

func TestPowerShellVerificationFailureDoesNotOverwriteWhenAvailable(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell install fixture requires Windows to execute yunxiao.exe")
	}
	pwsh := requirePowerShell(t)
	root := repoRoot(t)
	server := fixtureServer(t, "v1.2.3", map[string]releaseFixture{
		"v1.2.3": newZipRelease(t, "v1.2.3", windowsFixtureBinaryWithVerifyExit(t, 42)),
	})
	binDir := filepath.Join(t.TempDir(), "bin")
	target := filepath.Join(binDir, "yunxiao.exe")
	require.NoError(t, os.MkdirAll(binDir, 0o755))
	require.NoError(t, os.WriteFile(target, []byte("old-content"), 0o755))
	env := installerEnv(t, map[string]string{
		"USERPROFILE":                       t.TempDir(),
		"YUNXIAO_INSTALL_OS":                "windows",
		"YUNXIAO_INSTALL_ARCH":              "amd64",
		"YUNXIAO_INSTALL_DOWNLOAD_BASE_URL": server.URL,
	})

	stdout, stderr, code := runCommand(t, env, pwsh, "-NoProfile", "-File", filepath.Join(root, "scripts", "install.ps1"), "-Version", "v1.2.3", "-BinDir", binDir)
	require.Equal(t, 6, code, "stdout=%s stderr=%s", stdout, stderr)
	require.Contains(t, stderr, "failed verification before install")
	require.Equal(t, "old-content", readFile(t, target))
}

func TestPowerShellInstallerParseAndDryRunWhenAvailable(t *testing.T) {
	pwsh := requirePowerShell(t)
	root := repoRoot(t)
	script := filepath.Join(root, "scripts", "install.ps1")

	binDir := filepath.Join(t.TempDir(), "bin")
	env := installerEnv(t, map[string]string{
		"USERPROFILE":          t.TempDir(),
		"YUNXIAO_INSTALL_OS":   "windows",
		"YUNXIAO_INSTALL_ARCH": "arm64",
	})
	stdout, stderr, code := runCommand(t, env, pwsh, "-NoProfile", "-File", script, "-Version", "v1.2.3", "-BinDir", binDir, "-DryRun")
	require.Equal(t, 0, code, "stdout=%s stderr=%s", stdout, stderr)
	require.Contains(t, stdout, "os=windows")
	require.Contains(t, stdout, "arch=arm64")
	require.Contains(t, stdout, "asset=yunxiao_v1.2.3_windows_arm64.zip")
	require.Contains(t, stdout, "checksums_url=https://github.com/nick3/yunxiao-cli/releases/download/v1.2.3/checksums.txt")
}

func TestGoreleaserNamingContracts(t *testing.T) {
	root := repoRoot(t)
	stable := readFile(t, filepath.Join(root, ".goreleaser.yaml"))
	preview := readFile(t, filepath.Join(root, ".goreleaser.preview.yaml"))
	previewWorkflow := readFile(t, filepath.Join(root, ".github", "workflows", "preview-release.yml"))

	require.Contains(t, stable, "{{ .ProjectName }}_{{ .Tag }}_{{ .Os }}_{{ .Arch }}")
	require.Contains(t, stable, "wrap_in_directory: false")
	require.NotContains(t, stable, "PREVIEW_ASSET_VERSION")
	require.NotContains(t, stable, "PREVIEW_VERSION")
	require.Contains(t, preview, "{{ .ProjectName }}_preview.{{ .Env.PREVIEW_ASSET_VERSION }}_{{ .Os }}_{{ .Arch }}")
	require.Contains(t, preview, "wrap_in_directory: false")
	require.Contains(t, preview, "version_template: \"{{ .Env.PREVIEW_VERSION }}\"")
	require.Contains(t, previewWorkflow, "release --config .goreleaser.preview.yaml --snapshot --clean")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func normalizeSlashes(value string) string {
	return strings.ReplaceAll(value, "\\", "/")
}

func fixtureServer(t *testing.T, latest string, releases map[string]releaseFixture) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "repos/nick3/yunxiao-cli/releases/latest" {
			_, _ = fmt.Fprintf(w, `{"tag_name":"%s"}`, latest)
			return
		}

		parts := strings.Split(path, "/")
		if len(parts) != 6 || parts[0] != "nick3" || parts[1] != "yunxiao-cli" || parts[2] != "releases" || parts[3] != "download" {
			http.NotFound(w, r)
			return
		}
		fixture, ok := releases[parts[4]]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if parts[5] == "checksums.txt" {
			_, _ = fmt.Fprintf(w, "%s  %s\n", fixture.checksum, fixture.asset)
			return
		}
		if parts[5] == fixture.asset {
			_, _ = w.Write(fixture.archive)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)
	return server
}

func newTarRelease(t *testing.T, version string, binary string) releaseFixture {
	t.Helper()
	return newTarReleaseWithEntries(t, version, map[string]string{"yunxiao": binary})
}

func newTarReleaseWithEntries(t *testing.T, version string, entries map[string]string) releaseFixture {
	t.Helper()
	asset := fmt.Sprintf("yunxiao_%s_linux_amd64.tar.gz", version)
	archive := buildTarGz(t, entries)
	return releaseFixture{asset: asset, archive: archive, checksum: checksumHex(archive)}
}

func newZipRelease(t *testing.T, version string, binary []byte) releaseFixture {
	t.Helper()
	return newZipReleaseWithEntries(t, version, map[string][]byte{"yunxiao.exe": binary})
}

func newZipReleaseWithEntries(t *testing.T, version string, entries map[string][]byte) releaseFixture {
	t.Helper()
	asset := fmt.Sprintf("yunxiao_%s_windows_amd64.zip", version)
	archive := buildZip(t, entries)
	return releaseFixture{asset: asset, archive: archive, checksum: checksumHex(archive)}
}

func buildTarGz(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range entries {
		header := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(content))}
		require.NoError(t, tw.WriteHeader(header))
		_, err := tw.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return buf.Bytes()
}

func buildZip(t *testing.T, entries map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range entries {
		writer, err := zw.Create(name)
		require.NoError(t, err)
		_, err = writer.Write(content)
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

func fixtureShellBinary(marker string) string {
	return fmt.Sprintf("#!/bin/sh\nif [ \"$1\" = \"commands\" ] && [ \"$2\" = \"--json\" ]; then\n  printf '{\"version\":\"v1\",\"data\":[],\"meta\":{},\"error\":null}\\n'\n  exit 0\nfi\nprintf '%%s\\n' %q\n", marker)
}

func fixtureFailingVerifyBinary(marker string) string {
	return fmt.Sprintf("#!/bin/sh\nif [ \"$1\" = \"commands\" ] && [ \"$2\" = \"--json\" ]; then\n  printf 'verify failed\\n' >&2\n  exit 42\nfi\nprintf '%%s\\n' %q\n", marker)
}

func windowsFixtureBinaryWithMarker(t *testing.T, marker string) []byte {
	t.Helper()
	return windowsFixtureBinaryWithMarkerAndVerifyExit(t, marker, 0)
}

func windowsFixtureBinaryWithVerifyExit(t *testing.T, verifyExitCode int) []byte {
	t.Helper()
	return windowsFixtureBinaryWithMarkerAndVerifyExit(t, "ps-fixture-v1", verifyExitCode)
}

func windowsFixtureBinaryWithMarkerAndVerifyExit(t *testing.T, marker string, verifyExitCode int) []byte {
	t.Helper()
	sourceDir := t.TempDir()
	source := fmt.Sprintf(`package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) == 3 && os.Args[1] == "commands" && os.Args[2] == "--json" {
		fmt.Println(%q)
		os.Exit(%d)
	}
	fmt.Println(%q)
}
`, `{"version":"v1","data":[],"meta":{},"error":null}`, verifyExitCode, marker)
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "main.go"), []byte(source), 0o644))
	binaryPath := filepath.Join(sourceDir, "yunxiao.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, filepath.Join(sourceDir, "main.go"))
	cmd.WaitDelay = 250 * time.Millisecond
	cmd.Dir = sourceDir
	cmd.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=0")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	content, err := os.ReadFile(binaryPath)
	require.NoError(t, err)
	return content
}

func checksumHex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func requirePowerShell(t *testing.T) string {
	t.Helper()
	pwsh, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("pwsh is not available")
	}
	return pwsh
}

func runCommand(t *testing.T, env []string, name string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Env = env
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return stdout.String(), stderr.String(), 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return stdout.String(), stderr.String(), exitErr.ExitCode()
	}
	t.Fatalf("failed to run %s: %v", name, err)
	return "", "", 1
}

func installerEnv(t *testing.T, overrides map[string]string) []string {
	t.Helper()
	env := make([]string, 0, len(os.Environ())+len(overrides)+2)
	for _, item := range os.Environ() {
		if strings.HasPrefix(item, "YUNXIAO_") || strings.HasPrefix(item, "HOME=") || strings.HasPrefix(item, "TMPDIR=") || strings.HasPrefix(item, "USERPROFILE=") {
			continue
		}
		env = append(env, item)
	}
	if _, ok := overrides["HOME"]; !ok {
		env = append(env, "HOME="+t.TempDir())
	}
	if _, ok := overrides["TMPDIR"]; !ok {
		env = append(env, "TMPDIR="+t.TempDir())
	}
	for key, value := range overrides {
		env = append(env, key+"="+value)
	}
	return env
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}
