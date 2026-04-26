package org

import (
	"context"
	"fmt"
	"os"

	"github.com/aliyun/yunxiao-cli/internal/auth"
	"github.com/aliyun/yunxiao-cli/internal/cli"
	"github.com/aliyun/yunxiao-cli/internal/command/flagmeta"
	"github.com/aliyun/yunxiao-cli/internal/command/validation"
	"github.com/aliyun/yunxiao-cli/internal/config"
	orgdomain "github.com/aliyun/yunxiao-cli/internal/domains/org"
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
	cmd.AddCommand(newMembersCmd())
	return cmd
}

func newCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show current user and organization context",
		RunE:  runCurrent,
	}
}

func newMembersCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "members", Short: "Organization member collection commands"}
	cmd.AddCommand(newMembersListCmd())
	return cmd
}

func newMembersListCmd() *cobra.Command {
	var organizationID, pageToken string
	var pageSize int

	cmd := &cobra.Command{Use: "list", Short: "List organization members", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		populateTraceID(cmd, meta)
		orgID := config.GetOrganizationID(organizationID)
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		if errDetail := validation.PageSize(pageSize); errDetail != nil {
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, pagination, errDetail := orgdomain.ListMembers(context.Background(), client, orgID, pageSize, pageToken)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] member list failed: %s\n", errDetail.Message)
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		meta.Pagination = pagination
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	cmd.Flags().StringVar(&organizationID, "organization-id", "", "Organization ID")
	cmd.Flags().IntVar(&pageSize, "page-size", 100, "Page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	flagmeta.MustMarkRequired(cmd, "organization-id")
	return cmd
}

func runCurrent(cmd *cobra.Command, args []string) error {
	format := cli.GetOutputFormat()
	meta := &output.Meta{}
	client, ok := newAPIClient(cmd, format, meta)
	if !ok {
		return nil
	}

	currentUser, errDetail := orgdomain.GetCurrentUser(context.Background(), client)
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

	if code := cli.WriteResult(currentUser.Data, meta, format); code != cli.ExitSuccess {
		os.Exit(code)
	}
	return nil
}

func populateTraceID(cmd *cobra.Command, meta *output.Meta) {
	traceID, _ := cmd.Flags().GetString("trace-id")
	meta.TraceID = traceID
}

func newAPIClient(cmd *cobra.Command, format cli.OutputFormat, meta *output.Meta) (*httpx.Client, bool) {
	populateTraceID(cmd, meta)
	traceID := meta.TraceID
	timeoutFlag := cmd.Flags().Lookup("timeout")
	timeout, _ := cmd.Flags().GetInt("timeout")
	timeout = config.GetTimeout(timeout, timeoutFlag != nil && timeoutFlag.Changed)
	noRetry, _ := cmd.Flags().GetBool("no-retry")
	quiet, _ := cmd.Flags().GetBool("quiet")
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
