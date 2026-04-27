package projex

import (
	"context"
	"fmt"
	"os"

	"github.com/nick3/yunxiao-cli/internal/cli"
	"github.com/nick3/yunxiao-cli/internal/command/flagmeta"
	"github.com/nick3/yunxiao-cli/internal/config"
	projexdomain "github.com/nick3/yunxiao-cli/internal/domains/projex"
	"github.com/nick3/yunxiao-cli/internal/httpx"
	"github.com/nick3/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
)

func newWorkitemTypesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "workitem-types", Short: "Workitem type collection commands"}
	cmd.AddCommand(newWorkitemTypesListCmd())
	cmd.AddCommand(newWorkitemTypesRelationsCmd())
	return cmd
}

func newWorkitemTypeCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "workitem-type", Short: "Workitem type commands"}
	cmd.AddCommand(newWorkitemTypeGetCmd())
	cmd.AddCommand(newWorkitemTypeFieldsCmd())
	cmd.AddCommand(newWorkitemTypeWorkflowCmd())
	return cmd
}

func newWorkitemTypesListCmd() *cobra.Command {
	var organizationID, projectID, category string
	var all bool
	cmd := &cobra.Command{Use: "list", Short: "List workitem types", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(organizationID)
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		if all && projectID != "" {
			exitWithError(format, meta, "PARAM_INVALID", "param", false, "all cannot be used with project_id")
			return nil
		}
		if !all && projectID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "project_id is required unless all is set")
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		var data []map[string]any
		var errDetail *output.ErrorDetail
		if all {
			data, errDetail = projexdomain.ListAllWorkitemTypes(context.Background(), client, orgID)
		} else {
			data, errDetail = projexdomain.ListProjectWorkitemTypes(context.Background(), client, orgID, projectID, category)
		}
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] workitem type list failed: %s\n", errDetail.Message)
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	cmd.Flags().StringVar(&organizationID, "organization-id", "", "Organization ID")
	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID")
	cmd.Flags().StringVar(&category, "category", "", "Workitem category")
	cmd.Flags().BoolVar(&all, "all", false, "List all organization workitem types")
	flagmeta.MustMarkRequired(cmd, "organization-id")
	return cmd
}

func newWorkitemTypesRelationsCmd() *cobra.Command {
	var organizationID, typeID, relationType string
	cmd := &cobra.Command{Use: "relations", Short: "List related workitem types", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(organizationID)
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		if typeID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "workitem_type_id is required")
			return nil
		}
		if !validateRelationType(relationType, format, meta) {
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, errDetail := projexdomain.ListWorkitemRelationTypes(context.Background(), client, orgID, typeID, relationType)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] workitem type relations failed: %s\n", errDetail.Message)
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	cmd.Flags().StringVar(&organizationID, "organization-id", "", "Organization ID")
	cmd.Flags().StringVar(&typeID, "workitem-type-id", "", "Workitem type ID")
	cmd.Flags().StringVar(&relationType, "relation-type", "", "Relation type: PARENT, SUB, ASSOCIATED, DEPEND_ON, or DEPENDED_BY")
	flagmeta.MustMarkRequired(cmd, "organization-id", "workitem-type-id")
	return cmd
}

func newWorkitemTypeGetCmd() *cobra.Command {
	var organizationID, typeID string
	cmd := &cobra.Command{Use: "get", Short: "Get workitem type details", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(organizationID)
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		if typeID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "workitem_type_id is required")
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, errDetail := projexdomain.GetWorkitemType(context.Background(), client, orgID, typeID)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] workitem type lookup failed: %s\n", errDetail.Message)
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	cmd.Flags().StringVar(&organizationID, "organization-id", "", "Organization ID")
	cmd.Flags().StringVar(&typeID, "workitem-type-id", "", "Workitem type ID")
	flagmeta.MustMarkRequired(cmd, "organization-id", "workitem-type-id")
	return cmd
}

func newWorkitemTypeFieldsCmd() *cobra.Command {
	return newWorkitemTypeProjectMetadataCmd("fields", "Get workitem type field config", "[ERROR] workitem type fields failed: %s\n", projexdomain.GetWorkitemTypeFields)
}

func newWorkitemTypeWorkflowCmd() *cobra.Command {
	return newWorkitemTypeProjectMetadataCmd("workflow", "Get workitem type workflow", "[ERROR] workitem type workflow failed: %s\n", projexdomain.GetWorkitemTypeWorkflow)
}

func newWorkitemTypeProjectMetadataCmd(use, short, errorFormat string, call func(context.Context, *httpx.Client, string, string, string) (map[string]any, *output.ErrorDetail)) *cobra.Command {
	var organizationID, projectID, typeID string
	cmd := &cobra.Command{Use: use, Short: short, RunE: func(cmd *cobra.Command, args []string) error {
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
		if typeID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "workitem_type_id is required")
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, errDetail := call(context.Background(), client, orgID, projectID, typeID)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, errorFormat, errDetail.Message)
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	cmd.Flags().StringVar(&organizationID, "organization-id", "", "Organization ID")
	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID")
	cmd.Flags().StringVar(&typeID, "workitem-type-id", "", "Workitem type ID")
	flagmeta.MustMarkRequired(cmd, "organization-id", "project-id", "workitem-type-id")
	return cmd
}

func validateRelationType(value string, format cli.OutputFormat, meta *output.Meta) bool {
	switch value {
	case "", "PARENT", "SUB", "ASSOCIATED", "DEPEND_ON", "DEPENDED_BY":
		return true
	default:
		exitWithError(format, meta, "PARAM_INVALID", "param", false, "relation_type must be one of PARENT, SUB, ASSOCIATED, DEPEND_ON, or DEPENDED_BY")
		return false
	}
}
