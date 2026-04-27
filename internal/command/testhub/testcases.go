package testhub

import (
	"context"
	"fmt"
	"os"

	"github.com/nick3/yunxiao-cli/internal/auth"
	"github.com/nick3/yunxiao-cli/internal/cli"
	"github.com/nick3/yunxiao-cli/internal/command/flagmeta"
	"github.com/nick3/yunxiao-cli/internal/command/validation"
	"github.com/nick3/yunxiao-cli/internal/config"
	testhubdomain "github.com/nick3/yunxiao-cli/internal/domains/testhub"
	"github.com/nick3/yunxiao-cli/internal/httpx"
	"github.com/nick3/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
)

func NewTesthubCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "testhub", Short: "Testhub commands"}
	cmd.AddCommand(newTestcasesCmd())
	cmd.AddCommand(newTestcaseCmd())
	cmd.AddCommand(newDirectoriesCmd())
	cmd.AddCommand(newTestplansCmd())
	return cmd
}

func newTestcasesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "testcases", Short: "Testcase collection commands"}
	cmd.AddCommand(newTestcasesListCmd())
	return cmd
}

func newTestcaseCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "testcase", Short: "Testcase commands"}
	cmd.AddCommand(newTestcaseGetCmd())
	return cmd
}

func newTestcasesListCmd() *cobra.Command {
	var organizationID, testRepoID, pageToken string
	var pageSize int
	cmd := &cobra.Command{Use: "list", Short: "List testcases", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(organizationID)
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		if testRepoID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "test_repo_id is required")
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
		data, pagination, errDetail := testhubdomain.ListTestcases(context.Background(), client, orgID, testRepoID, pageSize, pageToken)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] testcase list failed: %s\n", errDetail.Message)
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
	cmd.Flags().StringVar(&testRepoID, "test-repo-id", "", "Test repository ID")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	flagmeta.MustMarkRequired(cmd, "organization-id", "test-repo-id")
	return cmd
}

func newTestcaseGetCmd() *cobra.Command {
	var organizationID, testRepoID, testcaseID string
	cmd := &cobra.Command{Use: "get", Short: "Get testcase details", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(organizationID)
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		if testRepoID == "" || testcaseID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "test_repo_id and testcase_id are required")
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, errDetail := testhubdomain.GetTestcase(context.Background(), client, orgID, testRepoID, testcaseID)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] testcase lookup failed: %s\n", errDetail.Message)
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	cmd.Flags().StringVar(&organizationID, "organization-id", "", "Organization ID")
	cmd.Flags().StringVar(&testRepoID, "test-repo-id", "", "Test repository ID")
	cmd.Flags().StringVar(&testcaseID, "testcase-id", "", "Testcase ID")
	flagmeta.MustMarkRequired(cmd, "organization-id", "test-repo-id", "testcase-id")
	return cmd
}

func newDirectoriesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "directories", Short: "Testhub directory collection commands"}
	cmd.AddCommand(newDirectoriesListCmd())
	return cmd
}

func newTestplansCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "testplans", Short: "Testplan collection commands"}
	cmd.AddCommand(newTestplansListCmd())
	return cmd
}

func newDirectoriesListCmd() *cobra.Command {
	var organizationID, testRepoID string
	cmd := &cobra.Command{Use: "list", Short: "List testhub directories", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(organizationID)
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		if testRepoID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "test_repo_id is required")
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, errDetail := testhubdomain.ListDirectories(context.Background(), client, orgID, testRepoID)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] directory list failed: %s\n", errDetail.Message)
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	cmd.Flags().StringVar(&organizationID, "organization-id", "", "Organization ID")
	cmd.Flags().StringVar(&testRepoID, "test-repo-id", "", "Test repository ID")
	flagmeta.MustMarkRequired(cmd, "organization-id", "test-repo-id")
	return cmd
}

func newTestplansListCmd() *cobra.Command {
	var organizationID string
	cmd := &cobra.Command{Use: "list", Short: "List testplans", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(organizationID)
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, errDetail := testhubdomain.ListTestplans(context.Background(), client, orgID)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] testplan list failed: %s\n", errDetail.Message)
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	cmd.Flags().StringVar(&organizationID, "organization-id", "", "Organization ID")
	flagmeta.MustMarkRequired(cmd, "organization-id")
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
