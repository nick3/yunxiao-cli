package codeup

import (
	"context"
	"fmt"
	"os"

	"github.com/nick3/yunxiao-cli/internal/auth"
	"github.com/nick3/yunxiao-cli/internal/cli"
	"github.com/nick3/yunxiao-cli/internal/command/flagmeta"
	"github.com/nick3/yunxiao-cli/internal/command/validation"
	"github.com/nick3/yunxiao-cli/internal/config"
	codeupdomain "github.com/nick3/yunxiao-cli/internal/domains/codeup"
	"github.com/nick3/yunxiao-cli/internal/httpx"
	"github.com/nick3/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
)

func NewCodeupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codeup",
		Short: "Codeup repository commands",
	}
	cmd.AddCommand(newReposCmd())
	cmd.AddCommand(newRepoCmd())
	cmd.AddCommand(newBranchesCmd())
	cmd.AddCommand(newCommitsCmd())
	cmd.AddCommand(newFileCmd())
	cmd.AddCommand(newCompareCmd())
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

func newBranchesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "branches", Short: "Branch collection commands"}
	cmd.AddCommand(newBranchesListCmd())
	return cmd
}

func newCommitsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "commits", Short: "Commit collection commands"}
	cmd.AddCommand(newCommitsListCmd())
	return cmd
}

func newFileCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "file", Short: "Repository file commands"}
	cmd.AddCommand(newFileGetCmd())
	return cmd
}

func newCompareCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "compare", Short: "Repository compare commands"}
	cmd.AddCommand(newCompareGetCmd())
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
			orgID, ok := requireOrganizationID(format, meta, organizationID)
			if !ok {
				return nil
			}
			if !validatePageSize(format, meta, pageSize) {
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
			orgID, ok := requireOrganizationID(format, meta, organizationID)
			if !ok {
				return nil
			}
			if !requireValue(format, meta, repoID, "repo_id is required") {
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

func newBranchesListCmd() *cobra.Command {
	var organizationID, repoID, pageToken, sort, search string
	var pageSize int

	cmd := &cobra.Command{Use: "list", Short: "List repository branches", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID, ok := requireOrganizationID(format, meta, organizationID)
		if !ok {
			return nil
		}
		if !requireValue(format, meta, repoID, "repo_id is required") || !validatePageSize(format, meta, pageSize) {
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, pagination, errDetail := codeupdomain.ListBranches(context.Background(), client, orgID, repoID, pageSize, pageToken, sort, search)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] branch list failed: %s\n", errDetail.Message)
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
	cmd.Flags().StringVar(&repoID, "repo-id", "", "Repository ID")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	cmd.Flags().StringVar(&sort, "sort", "", "Branch sort expression")
	cmd.Flags().StringVar(&search, "search", "", "Branch search keyword")
	flagmeta.MustMarkRequired(cmd, "organization-id", "repo-id")
	return cmd
}

func newCommitsListCmd() *cobra.Command {
	var organizationID, repoID, pageToken string
	var refName, since, until, filePath, search, committerIDs string
	var pageSize int
	var showSignature bool

	cmd := &cobra.Command{Use: "list", Short: "List repository commits", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID, ok := requireOrganizationID(format, meta, organizationID)
		if !ok {
			return nil
		}
		if !requireValue(format, meta, repoID, "repo_id is required") || !validatePageSize(format, meta, pageSize) {
			return nil
		}
		var showSignaturePtr *bool
		if cmd.Flags().Changed("show-signature") {
			showSignaturePtr = &showSignature
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		opts := codeupdomain.CommitListOptions{RefName: refName, Since: since, Until: until, Path: filePath, Search: search, CommitterIDs: committerIDs, ShowSignature: showSignaturePtr}
		data, pagination, errDetail := codeupdomain.ListCommits(context.Background(), client, orgID, repoID, pageSize, pageToken, opts)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] commit list failed: %s\n", errDetail.Message)
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
	cmd.Flags().StringVar(&repoID, "repo-id", "", "Repository ID")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	cmd.Flags().StringVar(&refName, "ref-name", "", "Branch, tag, or commit ref name")
	cmd.Flags().StringVar(&since, "since", "", "Lower commit time bound")
	cmd.Flags().StringVar(&until, "until", "", "Upper commit time bound")
	cmd.Flags().StringVar(&filePath, "path", "", "File path filter")
	cmd.Flags().StringVar(&search, "search", "", "Commit search keyword")
	cmd.Flags().StringVar(&committerIDs, "committer-ids", "", "Comma-separated committer IDs")
	cmd.Flags().BoolVar(&showSignature, "show-signature", false, "Include commit signature details")
	flagmeta.MustMarkRequired(cmd, "organization-id", "repo-id")
	return cmd
}

func newFileGetCmd() *cobra.Command {
	var organizationID, repoID, filePath, ref string

	cmd := &cobra.Command{Use: "get", Short: "Get repository file content", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID, ok := requireOrganizationID(format, meta, organizationID)
		if !ok {
			return nil
		}
		if !requireValue(format, meta, repoID, "repo_id is required") || !requireValue(format, meta, filePath, "path is required") || !requireValue(format, meta, ref, "ref is required") {
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, errDetail := codeupdomain.GetFile(context.Background(), client, orgID, repoID, filePath, ref)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] file lookup failed: %s\n", errDetail.Message)
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	cmd.Flags().StringVar(&organizationID, "organization-id", "", "Organization ID")
	cmd.Flags().StringVar(&repoID, "repo-id", "", "Repository ID")
	cmd.Flags().StringVar(&filePath, "path", "", "Repository file path")
	cmd.Flags().StringVar(&ref, "ref", "", "Branch, tag, or commit ref")
	flagmeta.MustMarkRequired(cmd, "organization-id", "repo-id", "path", "ref")
	return cmd
}

func newCompareGetCmd() *cobra.Command {
	var organizationID, repoID string
	var from, to, sourceType, targetType string
	var straight bool

	cmd := &cobra.Command{Use: "get", Short: "Compare two repository refs", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID, ok := requireOrganizationID(format, meta, organizationID)
		if !ok {
			return nil
		}
		if !requireValue(format, meta, repoID, "repo_id is required") || !requireValue(format, meta, from, "from is required") || !requireValue(format, meta, to, "to is required") {
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		var straightPtr *bool
		if cmd.Flags().Changed("straight") {
			straightPtr = &straight
		}
		opts := codeupdomain.CompareOptions{From: from, To: to, SourceType: sourceType, TargetType: targetType, Straight: straightPtr}
		data, errDetail := codeupdomain.GetCompare(context.Background(), client, orgID, repoID, opts)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] compare lookup failed: %s\n", errDetail.Message)
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	cmd.Flags().StringVar(&organizationID, "organization-id", "", "Organization ID")
	cmd.Flags().StringVar(&repoID, "repo-id", "", "Repository ID")
	cmd.Flags().StringVar(&from, "from", "", "Source ref")
	cmd.Flags().StringVar(&to, "to", "", "Target ref")
	cmd.Flags().StringVar(&sourceType, "source-type", "", "Source ref type")
	cmd.Flags().StringVar(&targetType, "target-type", "", "Target ref type")
	cmd.Flags().BoolVar(&straight, "straight", false, "Use straight comparison")
	flagmeta.MustMarkRequired(cmd, "organization-id", "repo-id", "from", "to")
	return cmd
}

func requireOrganizationID(format cli.OutputFormat, meta *output.Meta, organizationID string) (string, bool) {
	orgID := config.GetOrganizationID(organizationID)
	if orgID == "" {
		exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
		return "", false
	}
	return orgID, true
}

func requireValue(format cli.OutputFormat, meta *output.Meta, value, message string) bool {
	if value != "" {
		return true
	}
	exitWithError(format, meta, "PARAM_REQUIRED", "param", false, message)
	return false
}

func validatePageSize(format cli.OutputFormat, meta *output.Meta, pageSize int) bool {
	if errDetail := validation.PageSize(pageSize); errDetail != nil {
		os.Exit(cli.WriteError(errDetail, meta, format))
		return false
	}
	return true
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
