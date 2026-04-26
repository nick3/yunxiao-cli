package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aliyun/yunxiao-cli/internal/cli"
	"github.com/aliyun/yunxiao-cli/internal/config"
	"github.com/aliyun/yunxiao-cli/internal/domains/shared"
	"github.com/aliyun/yunxiao-cli/internal/httpx"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

type statusData struct {
	Source     string `json:"source"`
	Configured bool   `json:"configured"`
	Verified   *bool  `json:"verified"`
}

type loginData struct {
	Saved    bool   `json:"saved"`
	Source   string `json:"source"`
	Verified bool   `json:"verified"`
	Path     string `json:"path"`
}

type logoutData struct {
	Removed bool   `json:"removed"`
	Source  string `json:"source"`
}

func NewAuthCmd() *cobra.Command {
	var skipVerify bool
	var force bool
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(cmd, false, skipVerify, force)
		},
	}
	cmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "Save token without verifying it")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing config token")
	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newLogoutCmd())
	return cmd
}

func newLoginCmd() *cobra.Command {
	var tokenStdin bool
	var skipVerify bool
	var force bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Configure Yunxiao access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(cmd, tokenStdin, skipVerify, force)
		},
	}
	cmd.Flags().BoolVar(&tokenStdin, "token-stdin", false, "Read access token from stdin")
	cmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "Save token without verifying it")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing config token")
	return cmd
}

func runLogin(cmd *cobra.Command, tokenStdin, skipVerify, force bool) error {
	format := cli.GetOutputFormat()
	if !tokenStdin && !isTerminal(cmd.InOrStdin()) {
		exitWithError(format, "PARAM_INVALID", "param", "interactive login requires a terminal; use `yunxiao auth login --token-stdin` for scripts and CI")
		return nil
	}

	exists, err := hasConfigAccessToken()
	if err != nil {
		exitWithError(format, "CONFIG_WRITE_FAILED", "general", err.Error())
		return nil
	}
	if exists && !force {
		exitWithError(format, "PARAM_INVALID", "param", "config access token already exists; pass --force to overwrite")
		return nil
	}

	token, errDetail := readLoginToken(cmd.InOrStdin(), tokenStdin)
	if errDetail != nil {
		os.Exit(cli.WriteError(errDetail, &output.Meta{}, format))
		return nil
	}

	verified := false
	if !skipVerify {
		if errDetail := verifyToken(cmd, token); errDetail != nil {
			os.Exit(cli.WriteError(errDetail, &output.Meta{}, format))
			return nil
		}
		verified = true
	}

	path, err := writeAccessToken(token)
	if err != nil {
		exitWithError(format, "CONFIG_WRITE_FAILED", "general", err.Error())
		return nil
	}
	data := loginData{Saved: true, Source: "config", Verified: verified, Path: path}
	if code := cli.WriteResult(data, &output.Meta{}, format); code != cli.ExitSuccess {
		os.Exit(code)
	}
	return nil
}

func newStatusCmd() *cobra.Command {
	var verify bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := cli.GetOutputFormat()
			status, token := detectStatus()
			if verify {
				if !status.Configured {
					exitWithError(format, "AUTH_FAILED", "auth", "access token is not configured")
					return nil
				}
				if errDetail := verifyToken(cmd, token); errDetail != nil {
					os.Exit(cli.WriteError(errDetail, &output.Meta{}, format))
					return nil
				}
				verified := true
				status.Verified = &verified
			}
			if code := cli.WriteResult(status, &output.Meta{}, format); code != cli.ExitSuccess {
				os.Exit(code)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&verify, "verify", false, "Verify the current access token")
	return cmd
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove configured Yunxiao access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := cli.GetOutputFormat()
			removed, err := removeAccessToken()
			if err != nil {
				exitWithError(format, "CONFIG_WRITE_FAILED", "general", err.Error())
				return nil
			}
			data := logoutData{Removed: removed, Source: "config"}
			if code := cli.WriteResult(data, &output.Meta{}, format); code != cli.ExitSuccess {
				os.Exit(code)
			}
			return nil
		},
	}
}

func detectStatus() (statusData, string) {
	if token, ok := os.LookupEnv("YUNXIAO_ACCESS_TOKEN"); ok && token != "" {
		return statusData{Source: "env", Configured: true, Verified: nil}, token
	}
	if token := viper.GetString("access_token"); token != "" {
		return statusData{Source: "config", Configured: true, Verified: nil}, token
	}
	return statusData{Source: "none", Configured: false, Verified: nil}, ""
}

func verifyToken(cmd *cobra.Command, token string) *output.ErrorDetail {
	traceID, _ := cmd.Flags().GetString("trace-id")
	timeoutFlag := cmd.Flags().Lookup("timeout")
	timeout, _ := cmd.Flags().GetInt("timeout")
	timeout = config.GetTimeout(timeout, timeoutFlag != nil && timeoutFlag.Changed)
	noRetry, _ := cmd.Flags().GetBool("no-retry")
	quiet, _ := cmd.Flags().GetBool("quiet")

	client := httpx.NewClient(config.GetBaseURL(), token, timeout, noRetry, traceID)
	client.Quiet = quiet
	var data any
	return shared.RequestJSON(context.Background(), client, http.MethodGet, "/oapi/v1/platform/user", &data)
}

func readLoginToken(r io.Reader, tokenStdin bool) (string, *output.ErrorDetail) {
	var body string
	var err error
	if tokenStdin {
		var raw []byte
		raw, err = io.ReadAll(r)
		body = string(raw)
	} else {
		fmt.Fprint(os.Stderr, "Enter Yunxiao access token: ")
		body, err = bufio.NewReader(r).ReadString('\n')
		if errors.Is(err, io.EOF) {
			err = nil
		}
	}
	if err != nil {
		return "", &output.ErrorDetail{Code: "PARAM_INVALID", Category: "param", Retryable: false, Message: err.Error()}
	}
	return normalizeToken(body)
}

func normalizeToken(token string) (string, *output.ErrorDetail) {
	token = strings.TrimSuffix(token, "\n")
	token = strings.TrimSuffix(token, "\r")
	if token == "" {
		return "", &output.ErrorDetail{Code: "PARAM_INVALID", Category: "param", Retryable: false, Message: "access token is required"}
	}
	if strings.ContainsAny(token, "\r\n") {
		return "", &output.ErrorDetail{Code: "PARAM_INVALID", Category: "param", Retryable: false, Message: "access token must be a single line"}
	}
	return token, nil
}

func isTerminal(r io.Reader) bool {
	file, ok := r.(*os.File)
	return ok && term.IsTerminal(int(file.Fd()))
}

func hasConfigAccessToken() (bool, error) {
	path, err := configFilePath()
	if err != nil {
		return false, err
	}
	values, err := readConfigMap(path)
	if err != nil {
		return false, err
	}
	token, ok := values["access_token"].(string)
	return ok && token != "", nil
}

func writeAccessToken(token string) (string, error) {
	path, err := configFilePath()
	if err != nil {
		return "", err
	}
	values, err := readConfigMap(path)
	if err != nil {
		return "", err
	}
	values["access_token"] = token
	if err := writeConfigMap(path, values); err != nil {
		return "", err
	}
	return path, nil
}

func removeAccessToken() (bool, error) {
	path, err := configFilePath()
	if err != nil {
		return false, err
	}
	values, err := readConfigMap(path)
	if err != nil {
		return false, err
	}
	_, removed := values["access_token"]
	if !removed {
		return false, nil
	}
	delete(values, "access_token")
	return true, writeConfigMap(path, values)
}

func readConfigMap(path string) (map[string]any, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	if strings.TrimSpace(string(body)) == "" {
		return map[string]any{}, nil
	}
	values := map[string]any{}
	if err := yaml.Unmarshal(body, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func writeConfigMap(path string, values map[string]any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return err
	}
	body, err := yaml.Marshal(values)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*.yaml")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func configFilePath() (string, error) {
	if configDir := os.Getenv("YUNXIAO_CONFIG_FILE"); configDir != "" {
		return filepath.Join(configDir, "config.yaml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".yunxiao", "config.yaml"), nil
}

func exitWithError(format cli.OutputFormat, code, category, message string) {
	os.Exit(cli.WriteError(&output.ErrorDetail{Code: code, Category: category, Retryable: false, Message: message}, &output.Meta{}, format))
}
