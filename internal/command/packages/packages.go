package packages

import (
	"context"
	"fmt"
	"os"

	"github.com/nick3/yunxiao-cli/internal/auth"
	"github.com/nick3/yunxiao-cli/internal/cli"
	"github.com/nick3/yunxiao-cli/internal/command/flagmeta"
	"github.com/nick3/yunxiao-cli/internal/command/validation"
	"github.com/nick3/yunxiao-cli/internal/config"
	packagesdomain "github.com/nick3/yunxiao-cli/internal/domains/packages"
	"github.com/nick3/yunxiao-cli/internal/httpx"
	"github.com/nick3/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
)

func NewPackagesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "packages", Short: "Package repository commands"}
	cmd.AddCommand(newReposCmd())
	cmd.AddCommand(newArtifactsCmd())
	cmd.AddCommand(newArtifactCmd())
	return cmd
}

func newReposCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "repos", Short: "Package repository collection commands"}
	cmd.AddCommand(newReposListCmd())
	return cmd
}

func newArtifactsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "artifacts", Short: "Artifact collection commands"}
	cmd.AddCommand(newArtifactsListCmd())
	return cmd
}

func newArtifactCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "artifact", Short: "Artifact commands"}
	cmd.AddCommand(newArtifactGetCmd())
	return cmd
}

func newReposListCmd() *cobra.Command {
	var organizationID, repoTypes, repoCategories, pageToken string
	var pageSize int
	cmd := &cobra.Command{Use: "list", Short: "List package repositories", RunE: func(cmd *cobra.Command, args []string) error {
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
		data, pagination, errDetail := packagesdomain.ListRepositories(context.Background(), client, orgID, pageSize, pageToken, repoTypes, repoCategories)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] package repository list failed: %s\n", errDetail.Message)
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
	cmd.Flags().StringVar(&repoTypes, "repo-types", "", "Repository types")
	cmd.Flags().StringVar(&repoCategories, "repo-categories", "", "Repository categories")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	flagmeta.MustMarkRequired(cmd, "organization-id")
	return cmd
}

func newArtifactsListCmd() *cobra.Command {
	var organizationID, repoID, repoType, pageToken, search string
	var pageSize int
	cmd := &cobra.Command{Use: "list", Short: "List artifacts", RunE: func(cmd *cobra.Command, args []string) error {
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
		if repoType == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "repo_type is required")
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
		data, pagination, errDetail := packagesdomain.ListArtifacts(context.Background(), client, orgID, repoID, repoType, pageSize, pageToken, search)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] artifact list failed: %s\n", errDetail.Message)
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
	cmd.Flags().StringVar(&repoType, "repo-type", "", "Repository type")
	cmd.Flags().StringVar(&search, "search", "", "Search keyword")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	flagmeta.MustMarkRequired(cmd, "organization-id", "repo-id", "repo-type")
	return cmd
}

func newArtifactGetCmd() *cobra.Command {
	var organizationID, repoID, artifactID, repoType string
	cmd := &cobra.Command{Use: "get", Short: "Get artifact details", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(organizationID)
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		if repoID == "" || artifactID == "" || repoType == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "repo_id, artifact_id and repo_type are required")
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, errDetail := packagesdomain.GetArtifact(context.Background(), client, orgID, repoID, artifactID, repoType)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] artifact lookup failed: %s\n", errDetail.Message)
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
	cmd.Flags().StringVar(&artifactID, "artifact-id", "", "Artifact ID")
	cmd.Flags().StringVar(&repoType, "repo-type", "", "Repository type")
	flagmeta.MustMarkRequired(cmd, "organization-id", "repo-id", "artifact-id", "repo-type")
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
