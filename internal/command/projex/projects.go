package projex

import (
	"context"
	"fmt"
	"os"

	"github.com/aliyun/yunxiao-cli/internal/auth"
	"github.com/aliyun/yunxiao-cli/internal/cli"
	"github.com/aliyun/yunxiao-cli/internal/command/flagmeta"
	"github.com/aliyun/yunxiao-cli/internal/command/validation"
	"github.com/aliyun/yunxiao-cli/internal/config"
	projexdomain "github.com/aliyun/yunxiao-cli/internal/domains/projex"
	"github.com/aliyun/yunxiao-cli/internal/httpx"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
)

func NewProjexCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "projex", Short: "Projex project commands"}
	cmd.AddCommand(newProjectsCmd())
	cmd.AddCommand(newProjectCmd())
	cmd.AddCommand(newWorkitemsCmd())
	cmd.AddCommand(newWorkitemCmd())
	cmd.AddCommand(newSprintsCmd())
	return cmd
}

func newProjectsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "projects", Short: "Project collection commands"}
	cmd.AddCommand(newProjectsListCmd())
	return cmd
}

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "project", Short: "Project commands"}
	cmd.AddCommand(newProjectGetCmd())
	return cmd
}

func newProjectsListCmd() *cobra.Command {
	var organizationID string
	var pageSize int
	var pageToken string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
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
			data, pagination, errDetail := projexdomain.ListProjects(context.Background(), client, orgID, pageSize, pageToken)
			if errDetail != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] project list failed: %s\n", errDetail.Message)
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
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	flagmeta.MustMarkRequired(cmd, "organization-id")
	return cmd
}

func newProjectGetCmd() *cobra.Command {
	var organizationID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get project details",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := cli.GetOutputFormat()
			meta := &output.Meta{}
			orgID := config.GetOrganizationID(organizationID)
			if orgID == "" {
				exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
				return nil
			}
			if projectID == "" {
				exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "project_id is required")
				return nil
			}
			client, ok := newAPIClient(cmd, format, meta)
			if !ok {
				return nil
			}
			data, errDetail := projexdomain.GetProject(context.Background(), client, orgID, projectID)
			if errDetail != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] project lookup failed: %s\n", errDetail.Message)
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
	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID")
	flagmeta.MustMarkRequired(cmd, "organization-id", "project-id")
	return cmd
}

func newWorkitemsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "workitems", Short: "Workitem collection commands"}
	cmd.AddCommand(newWorkitemsListCmd())
	return cmd
}

func newWorkitemCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "workitem", Short: "Workitem commands"}
	cmd.AddCommand(newWorkitemGetCmd())
	return cmd
}

func newSprintsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "sprints", Short: "Sprint collection commands"}
	cmd.AddCommand(newSprintsListCmd())
	return cmd
}

func newWorkitemsListCmd() *cobra.Command {
	var organizationID, category, spaceID, pageToken string
	var pageSize int

	cmd := &cobra.Command{Use: "list", Short: "List workitems", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(organizationID)
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		if category == "" || spaceID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "category and space_id are required")
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
		data, pagination, errDetail := projexdomain.ListWorkitems(context.Background(), client, orgID, category, spaceID, pageSize, pageToken)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] workitem list failed: %s\n", errDetail.Message)
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
	cmd.Flags().StringVar(&category, "category", "", "Workitem category")
	cmd.Flags().StringVar(&spaceID, "space-id", "", "Projex space ID")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	flagmeta.MustMarkRequired(cmd, "organization-id", "category", "space-id")
	return cmd
}

func newWorkitemGetCmd() *cobra.Command {
	var organizationID, workitemID string

	cmd := &cobra.Command{Use: "get", Short: "Get workitem details", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(organizationID)
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		if workitemID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "workitem_id is required")
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, errDetail := projexdomain.GetWorkitem(context.Background(), client, orgID, workitemID)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] workitem lookup failed: %s\n", errDetail.Message)
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	cmd.Flags().StringVar(&organizationID, "organization-id", "", "Organization ID")
	cmd.Flags().StringVar(&workitemID, "workitem-id", "", "Workitem ID")
	flagmeta.MustMarkRequired(cmd, "organization-id", "workitem-id")
	return cmd
}

func newSprintsListCmd() *cobra.Command {
	var organizationID, projectID, pageToken, status string
	var pageSize int

	cmd := &cobra.Command{Use: "list", Short: "List sprints", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(organizationID)
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		if projectID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "project_id is required")
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
		data, pagination, errDetail := projexdomain.ListSprints(context.Background(), client, orgID, projectID, pageSize, pageToken, status)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] sprint list failed: %s\n", errDetail.Message)
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
	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	cmd.Flags().StringVar(&status, "status", "", "Comma-separated sprint statuses")
	flagmeta.MustMarkRequired(cmd, "organization-id", "project-id")
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
