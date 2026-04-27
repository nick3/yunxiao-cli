package raw

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nick3/yunxiao-cli/internal/auth"
	"github.com/nick3/yunxiao-cli/internal/cli"
	"github.com/nick3/yunxiao-cli/internal/command/flagmeta"
	"github.com/nick3/yunxiao-cli/internal/config"
	rawdomain "github.com/nick3/yunxiao-cli/internal/domains/raw"
	"github.com/nick3/yunxiao-cli/internal/httpx"
	"github.com/nick3/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
)

func NewRawCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "raw", Short: "Raw Yunxiao API commands"}
	cmd.AddCommand(newRequestCmd())
	return cmd
}

func newRequestCmd() *cobra.Command {
	var method string
	var path string
	cmd := &cobra.Command{Use: "request", Short: "Send a read-only raw Yunxiao API request", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		method = strings.ToUpper(method)
		if method == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "method is required")
			return nil
		}
		if path == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "path is required")
			return nil
		}
		if errDetail := rawdomain.Validate(method, path); errDetail != nil {
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, errDetail := rawdomain.Request(context.Background(), client, method, path)
		if errDetail != nil {
			if errDetail.Category != "param" {
				fmt.Fprintf(os.Stderr, "[ERROR] raw request failed: %s\n", errDetail.Message)
			}
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	cmd.Flags().StringVar(&method, "method", "GET", "HTTP method; Phase 2 supports GET only")
	cmd.Flags().StringVar(&path, "path", "", "API path starting with /oapi/")
	flagmeta.MustMarkRequired(cmd, "method", "path")
	return cmd
}

func newAPIClient(cmd *cobra.Command, format cli.OutputFormat, meta *output.Meta) (*httpx.Client, bool) {
	traceID, _ := cmd.Flags().GetString("trace-id")
	timeoutFlag := cmd.Flags().Lookup("timeout")
	timeout, _ := cmd.Flags().GetInt("timeout")
	timeout = config.GetTimeout(timeout, timeoutFlag != nil && timeoutFlag.Changed)
	noRetry, _ := cmd.Flags().GetBool("no-retry")
	quiet, _ := cmd.Flags().GetBool("quiet")
	meta.TraceID = traceID
	token, err := auth.GetAccessToken()
	if err != nil {
		exitWithError(format, meta, "AUTH_FAILED", "auth", false, err.Error())
		return nil, false
	}
	client := httpx.NewClient(config.GetBaseURL(), token, timeout, noRetry, traceID)
	client.Quiet = quiet
	return client, true
}

func exitWithError(format cli.OutputFormat, meta *output.Meta, code, category string, retryable bool, message string) {
	os.Exit(cli.WriteError(&output.ErrorDetail{Code: code, Category: category, Retryable: retryable, Message: message}, meta, format))
}
