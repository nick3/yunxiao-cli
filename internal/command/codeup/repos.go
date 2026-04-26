package codeup

import (
	"context"
	"fmt"
	"os"

	"github.com/aliyun/yunxiao-cli/internal/auth"
	"github.com/aliyun/yunxiao-cli/internal/cli"
	"github.com/aliyun/yunxiao-cli/internal/command/flagmeta"
	"github.com/aliyun/yunxiao-cli/internal/command/validation"
	"github.com/aliyun/yunxiao-cli/internal/config"
	codeupdomain "github.com/aliyun/yunxiao-cli/internal/domains/codeup"
	"github.com/aliyun/yunxiao-cli/internal/httpx"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
)

func NewCodeupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codeup",
		Short: "Codeup repository commands",
	}
	cmd.AddCommand(newReposCmd())
	cmd.AddCommand(newRepoCmd())
	return cmd
}

func newReposCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repos",
		Short: "Repository collection commands",
	}
	cmd.AddCommand(newReposListCmd())
	return cmd
}

func newRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Repository commands",
	}
	cmd.AddCommand(newRepoGetCmd())
	return cmd
}

func newReposListCmd() *cobra.Command {
	var organizationID string
	var pageSize int
	var pageToken string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := cli.GetOutputFormat()
			meta := &output.Meta{}
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
			data, pagination, errDetail := codeupdomain.ListRepositories(context.Background(), client, orgID, pageSize, pageToken)
			if errDetail != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] repository list failed: %s\n", errDetail.Message)
				os.Exit(cli.WriteError(errDetail, meta, format))
				return nil
			}
			meta.Pagination = pagination
			if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
				os.Exit(code)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&organizationID, "organization-id", "", "Organization ID")
	cmd.Flags().IntVar(&pageSize, "page-size", 2, "Page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	flagmeta.MustMarkRequired(cmd, "organization-id")
	return cmd
}

func newRepoGetCmd() *cobra.Command {
	var organizationID string
	var repoID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get repository details",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := cli.GetOutputFormat()
			meta := &output.Meta{}
			orgID := config.GetOrganizationID(organizationID)
			if orgID == "" {
				exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
				return nil
			}
			if repoID == "" {
				exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "repo_id is required")
				return nil
			}
			client, ok := newAPIClient(cmd, format, meta)
			if !ok {
				return nil
			}
			data, errDetail := codeupdomain.GetRepository(context.Background(), client, orgID, repoID)
			if errDetail != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] repository lookup failed: %s\n", errDetail.Message)
				os.Exit(cli.WriteError(errDetail, meta, format))
				return nil
			}
			if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
				os.Exit(code)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&organizationID, "organization-id", "", "Organization ID")
	cmd.Flags().StringVar(&repoID, "repo-id", "", "Repository ID")
	flagmeta.MustMarkRequired(cmd, "organization-id", "repo-id")
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
