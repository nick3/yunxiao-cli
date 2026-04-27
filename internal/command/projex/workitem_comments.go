package projex

import (
	"context"
	"fmt"
	"os"

	"github.com/nick3/yunxiao-cli/internal/cli"
	"github.com/nick3/yunxiao-cli/internal/command/flagmeta"
	"github.com/nick3/yunxiao-cli/internal/command/validation"
	"github.com/nick3/yunxiao-cli/internal/config"
	projexdomain "github.com/nick3/yunxiao-cli/internal/domains/projex"
	"github.com/nick3/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
)

func newWorkitemCommentsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "comments", Short: "Workitem comment collection commands"}
	cmd.AddCommand(newWorkitemCommentsListCmd())
	return cmd
}

func newWorkitemCommentCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "comment", Short: "Workitem comment commands"}
	cmd.AddCommand(newWorkitemCommentCreateCmd())
	return cmd
}

func newWorkitemCommentsListCmd() *cobra.Command {
	var organizationID, workitemID, pageToken string
	var pageSize int
	cmd := &cobra.Command{Use: "list", Short: "List workitem comments", RunE: func(cmd *cobra.Command, args []string) error {
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
		if errDetail := validation.PageSize(pageSize); errDetail != nil {
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, pagination, errDetail := projexdomain.ListWorkitemComments(context.Background(), client, orgID, workitemID, pageSize, pageToken)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] workitem comments list failed: %s\n", errDetail.Message)
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
	cmd.Flags().StringVar(&workitemID, "workitem-id", "", "Workitem ID")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	flagmeta.MustMarkRequired(cmd, "organization-id", "workitem-id")
	return cmd
}

func newWorkitemCommentCreateCmd() *cobra.Command {
	var organizationID, workitemID, content, contentFile string
	var yes bool
	cmd := &cobra.Command{Use: "create", Short: "Create a workitem comment", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(organizationID)
		if !yes {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "yes is required because creating a workitem comment writes to Yunxiao")
			return nil
		}
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		if workitemID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "workitem_id is required")
			return nil
		}
		body, ok := resolveTextValue(cmd, "content", content, contentFile, "content", format, meta)
		if !ok {
			return nil
		}
		if body == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "content is required")
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, errDetail := projexdomain.CreateWorkitemComment(context.Background(), client, orgID, workitemID, body)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] workitem comment create failed: %s\n", errDetail.Message)
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
	cmd.Flags().StringVar(&content, "content", "", "Comment content")
	cmd.Flags().StringVar(&contentFile, "content-file", "", "Read comment content from a UTF-8 text file")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm this write operation")
	flagmeta.MustMarkRequired(cmd, "organization-id", "workitem-id", "yes")
	return cmd
}
