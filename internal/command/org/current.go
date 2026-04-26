package org

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/aliyun/yunxiao-cli/internal/auth"
	"github.com/aliyun/yunxiao-cli/internal/cli"
	"github.com/aliyun/yunxiao-cli/internal/config"
	"github.com/aliyun/yunxiao-cli/internal/domains/shared"
	"github.com/aliyun/yunxiao-cli/internal/httpx"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
)

func NewOrgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Organization management commands",
	}
	cmd.AddCommand(newCurrentCmd())
	return cmd
}

func newCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show current user and organization context",
		RunE:  runCurrent,
	}
}

func runCurrent(cmd *cobra.Command, args []string) error {
	format := cli.GetOutputFormat()
	traceID, _ := cmd.Flags().GetString("trace-id")
	timeoutFlag := cmd.Flags().Lookup("timeout")
	timeout, _ := cmd.Flags().GetInt("timeout")
	timeout = config.GetTimeout(timeout, timeoutFlag != nil && timeoutFlag.Changed)
	noRetry, _ := cmd.Flags().GetBool("no-retry")

	meta := &output.Meta{TraceID: traceID}

	token, err := auth.GetAccessToken()
	if err != nil {
		code := cli.WriteError(&output.ErrorDetail{
			Code: "AUTH_FAILED", Category: "auth",
			Retryable: false, Message: err.Error(),
		}, meta, format)
		os.Exit(code)
		return nil
	}

	baseURL := config.GetBaseURL()
	client := httpx.NewClient(baseURL, token, timeout, noRetry, traceID)
	quiet, _ := cmd.Flags().GetBool("quiet")
	client.Quiet = quiet

	var data any
	errDetail := shared.RequestJSON(context.Background(), client, http.MethodGet, "/oapi/v1/platform/user", &data)
	if errDetail != nil {
		if errDetail.Code == "RESPONSE_DECODE_FAILED" {
			fmt.Fprintf(os.Stderr, "[ERROR] org current response decode failed: %s\n", errDetail.Message)
		} else {
			fmt.Fprintf(os.Stderr, "[ERROR] org current failed: %s\n", errDetail.Message)
		}
		code := cli.WriteError(errDetail, meta, format)
		os.Exit(code)
		return nil
	}

	if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
		os.Exit(code)
	}
	return nil
}
