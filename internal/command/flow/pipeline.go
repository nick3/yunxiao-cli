package flow

import (
	"context"
	"fmt"
	"os"

	"github.com/aliyun/yunxiao-cli/internal/auth"
	"github.com/aliyun/yunxiao-cli/internal/cli"
	"github.com/aliyun/yunxiao-cli/internal/command/flagmeta"
	"github.com/aliyun/yunxiao-cli/internal/command/validation"
	"github.com/aliyun/yunxiao-cli/internal/config"
	flowdomain "github.com/aliyun/yunxiao-cli/internal/domains/flow"
	"github.com/aliyun/yunxiao-cli/internal/httpx"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
)

func NewFlowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flow",
		Short: "Pipeline and flow commands",
	}
	cmd.AddCommand(newPipelineCmd())
	cmd.AddCommand(newPipelinesCmd())
	cmd.AddCommand(newRunsCmd())
	cmd.AddCommand(newRunCmd())
	return cmd
}

func newPipelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Pipeline commands",
	}
	cmd.AddCommand(newPipelineGetCmd())
	return cmd
}

func newPipelinesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipelines",
		Short: "Pipeline collection commands",
	}
	cmd.AddCommand(newPipelinesListCmd())
	return cmd
}

func newRunsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "runs", Short: "Pipeline run collection commands"}
	cmd.AddCommand(newRunsListCmd())
	return cmd
}

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "run", Short: "Pipeline run commands"}
	cmd.AddCommand(newRunGetCmd())
	return cmd
}

func newPipelineGetCmd() *cobra.Command {
	var organizationID string
	var pipelineID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get pipeline details",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := cli.GetOutputFormat()
			meta := &output.Meta{}
			orgID, ok := requireOrganizationID(format, meta, organizationID)
			if !ok {
				return nil
			}
			if !requireValue(format, meta, pipelineID, "pipeline_id is required") {
				return nil
			}
			client, ok := newAPIClient(cmd, format, meta)
			if !ok {
				return nil
			}
			data, errDetail := flowdomain.GetPipeline(context.Background(), client, orgID, pipelineID)
			if errDetail != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] pipeline lookup failed: %s\n", errDetail.Message)
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
	cmd.Flags().StringVar(&pipelineID, "pipeline-id", "", "Pipeline ID")
	flagmeta.MustMarkRequired(cmd, "organization-id", "pipeline-id")
	return cmd
}

func newPipelinesListCmd() *cobra.Command {
	var organizationID string
	var pageSize int
	var pageToken string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pipelines",
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
			data, pagination, errDetail := flowdomain.ListPipelines(context.Background(), client, orgID, pageSize, pageToken)
			if errDetail != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] pipeline list failed: %s\n", errDetail.Message)
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

func newRunsListCmd() *cobra.Command {
	var organizationID, pipelineID, pageToken string
	var startTime, endTime, status, triggerMode string
	var pageSize int

	cmd := &cobra.Command{Use: "list", Short: "List pipeline runs", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID, ok := requireOrganizationID(format, meta, organizationID)
		if !ok {
			return nil
		}
		if !requireValue(format, meta, pipelineID, "pipeline_id is required") || !validatePageSize(format, meta, pageSize) {
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		opts := flowdomain.RunListOptions{StartTime: startTime, EndTime: endTime, Status: status, TriggerMode: triggerMode}
		data, pagination, errDetail := flowdomain.ListRuns(context.Background(), client, orgID, pipelineID, pageSize, pageToken, opts)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] run list failed: %s\n", errDetail.Message)
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
	cmd.Flags().StringVar(&pipelineID, "pipeline-id", "", "Pipeline ID")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	cmd.Flags().StringVar(&startTime, "start-time", "", "Run start time lower bound")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Run end time upper bound")
	cmd.Flags().StringVar(&status, "status", "", "Run status filter")
	cmd.Flags().StringVar(&triggerMode, "trigger-mode", "", "Run trigger mode filter")
	flagmeta.MustMarkRequired(cmd, "organization-id", "pipeline-id")
	return cmd
}

func newRunGetCmd() *cobra.Command {
	var organizationID, pipelineID, runID string

	cmd := &cobra.Command{Use: "get", Short: "Get pipeline run details", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID, ok := requireOrganizationID(format, meta, organizationID)
		if !ok {
			return nil
		}
		if !requireValue(format, meta, pipelineID, "pipeline_id is required") || !requireValue(format, meta, runID, "run_id is required") {
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		data, errDetail := flowdomain.GetRun(context.Background(), client, orgID, pipelineID, runID)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] run lookup failed: %s\n", errDetail.Message)
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	cmd.Flags().StringVar(&organizationID, "organization-id", "", "Organization ID")
	cmd.Flags().StringVar(&pipelineID, "pipeline-id", "", "Pipeline ID")
	cmd.Flags().StringVar(&runID, "run-id", "", "Pipeline run ID")
	flagmeta.MustMarkRequired(cmd, "organization-id", "pipeline-id", "run-id")
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
