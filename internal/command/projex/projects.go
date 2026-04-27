package projex

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/nick3/yunxiao-cli/internal/auth"
	"github.com/nick3/yunxiao-cli/internal/cli"
	"github.com/nick3/yunxiao-cli/internal/command/flagmeta"
	"github.com/nick3/yunxiao-cli/internal/command/validation"
	"github.com/nick3/yunxiao-cli/internal/config"
	orgdomain "github.com/nick3/yunxiao-cli/internal/domains/org"
	projexdomain "github.com/nick3/yunxiao-cli/internal/domains/projex"
	"github.com/nick3/yunxiao-cli/internal/httpx"
	"github.com/nick3/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
)

const maxAggregatePages = 10000

func NewProjexCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "projex", Short: "Projex project commands"}
	cmd.AddCommand(newProjectsCmd())
	cmd.AddCommand(newProjectCmd())
	cmd.AddCommand(newWorkitemsCmd())
	cmd.AddCommand(newWorkitemCmd())
	cmd.AddCommand(newWorkitemTypesCmd())
	cmd.AddCommand(newWorkitemTypeCmd())
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
	var mine bool
	var opts projexdomain.ProjectListOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := cli.GetOutputFormat()
			meta := &output.Meta{}
			if errDetail := validation.PageSize(pageSize); errDetail != nil {
				os.Exit(cli.WriteError(errDetail, meta, format))
				return nil
			}
			if mine && (opts.ScenarioFilter != "" || opts.UserID != "") {
				exitWithError(format, meta, "PARAM_INVALID", "param", false, "mine cannot be used with scenario_filter or user_id")
				return nil
			}
			if mine {
				opts.ScenarioFilter = "participate"
				opts.UserID = "self"
			}
			if opts.UserID != "" && opts.ScenarioFilter == "" {
				exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "scenario_filter is required when user_id is set")
				return nil
			}
			if !validProjectScenarioFilter(opts.ScenarioFilter) {
				exitWithError(format, meta, "PARAM_INVALID", "param", false, "scenario_filter must be one of manage, participate, or favorite")
				return nil
			}
			if opts.ScenarioFilter != "" && opts.UserID == "" {
				exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "user_id is required when scenario_filter is set")
				return nil
			}
			client, ok := newAPIClient(cmd, format, meta)
			if !ok {
				return nil
			}
			orgID, errDetail := resolveProjexProjectListContext(context.Background(), client, organizationID, &opts)
			if errDetail != nil {
				os.Exit(cli.WriteError(errDetail, meta, format))
				return nil
			}
			data, pagination, errDetail := projexdomain.ListProjects(context.Background(), client, orgID, pageSize, pageToken, opts)
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
	cmd.Flags().StringVar(&opts.Name, "name", "", "Project name keyword filter")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Comma-separated project status IDs")
	cmd.Flags().StringVar(&opts.CreatedAfter, "created-after", "", "Creation date lower bound")
	cmd.Flags().StringVar(&opts.CreatedBefore, "created-before", "", "Creation date upper bound")
	cmd.Flags().StringVar(&opts.Creator, "creator", "", "Comma-separated creator IDs")
	cmd.Flags().StringVar(&opts.AdminUserID, "admin-user-id", "", "Comma-separated project administrator IDs")
	cmd.Flags().StringVar(&opts.LogicalStatus, "logical-status", "", "Project logical status")
	cmd.Flags().StringVar(&opts.AdvancedConditions, "advanced-conditions", "", "Raw Projex conditions JSON")
	cmd.Flags().StringVar(&opts.ExtraConditions, "extra-conditions", "", "Raw Projex extraConditions JSON")
	cmd.Flags().StringVar(&opts.OrderBy, "order-by", "", "Project order field")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", "Sort direction")
	cmd.Flags().StringVar(&opts.ScenarioFilter, "scenario-filter", "", "Project scenario filter: manage, participate, or favorite")
	cmd.Flags().StringVar(&opts.UserID, "user-id", "", "User ID for scenario filter, or self for current user")
	cmd.Flags().BoolVar(&mine, "mine", false, "List projects participated in by the current user")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	return cmd
}

func validProjectScenarioFilter(value string) bool {
	switch value {
	case "", "manage", "participate", "favorite":
		return true
	default:
		return false
	}
}

func resolveProjexProjectListContext(ctx context.Context, client *httpx.Client, flagValue string, opts *projexdomain.ProjectListOptions) (string, *output.ErrorDetail) {
	orgID := config.GetOrganizationID(flagValue)
	needsCurrentUser := (!config.IsRegionBaseURL(client.BaseURL) && orgID == "") || opts.UserID == "self"
	if !needsCurrentUser {
		return orgID, nil
	}

	currentUser, errDetail := orgdomain.GetCurrentUser(ctx, client)
	if errDetail != nil {
		return "", errDetail
	}
	if opts.UserID == "self" {
		if currentUser.UserID == "" {
			return "", &output.ErrorDetail{Code: "PARAM_REQUIRED", Category: "param", Retryable: false, Message: "user_id is required because current user response has no id"}
		}
		opts.UserID = currentUser.UserID
	}
	if orgID == "" && !config.IsRegionBaseURL(client.BaseURL) {
		if currentUser.LastOrganizationID == "" {
			return "", &output.ErrorDetail{Code: "PARAM_REQUIRED", Category: "param", Retryable: false, Message: "organization_id is required because current user has no lastOrganization"}
		}
		orgID = currentUser.LastOrganizationID
	}
	return orgID, nil
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
	cmd.AddCommand(newWorkitemCreateCmd())
	cmd.AddCommand(newWorkitemUpdateCmd())
	cmd.AddCommand(newWorkitemCommentsCmd())
	cmd.AddCommand(newWorkitemCommentCmd())
	return cmd
}

func newSprintsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "sprints", Short: "Sprint collection commands"}
	cmd.AddCommand(newSprintsListCmd())
	return cmd
}

func newWorkitemsListCmd() *cobra.Command {
	var organizationID, category, projectID, spaceID, pageToken string
	var opts projexdomain.WorkitemListOptions
	var pageSize int
	var mine, unfinished bool

	cmd := &cobra.Command{Use: "list", Short: "List workitems", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(organizationID)
		if category == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "category is required")
			return nil
		}
		if unfinished && !mine {
			exitWithError(format, meta, "PARAM_INVALID", "param", false, "unfinished can only be used with mine")
			return nil
		}
		if mine && (projectID != "" || spaceID != "") {
			exitWithError(format, meta, "PARAM_INVALID", "param", false, "project_id or space_id cannot be used with mine")
			return nil
		}
		resolvedProjectID, ok := resolveProjexProjectID(projectID, spaceID, format, meta)
		if !ok {
			return nil
		}
		if !mine && resolvedProjectID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "project_id or space_id is required unless mine is set")
			return nil
		}
		if mine && pageToken != "" {
			exitWithError(format, meta, "PARAM_INVALID", "param", false, "page_token cannot be used with mine")
			return nil
		}
		if mine && opts.AssignedTo != "" {
			exitWithError(format, meta, "PARAM_INVALID", "param", false, "assigned_to cannot be used with mine")
			return nil
		}
		if mine && opts.AdvancedConditions != "" {
			exitWithError(format, meta, "PARAM_INVALID", "param", false, "advanced_conditions cannot be used with mine")
			return nil
		}
		if !mine && orgID == "" {
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
		if mine {
			data, pagination, errDetail := listMyWorkitems(context.Background(), client, orgID, category, pageSize, opts, unfinished)
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
		}
		data, pagination, errDetail := projexdomain.ListWorkitems(context.Background(), client, orgID, category, resolvedProjectID, pageSize, pageToken, opts)
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
	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID; alias of space-id")
	cmd.Flags().StringVar(&spaceID, "space-id", "", "Projex space ID; alias of project-id")
	cmd.Flags().StringVar(&opts.SpaceType, "space-type", "", "Projex space type")
	cmd.Flags().StringVar(&opts.Subject, "subject", "", "Subject keyword filter")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Comma-separated status IDs")
	cmd.Flags().StringVar(&opts.CreatedAfter, "created-after", "", "Creation date lower bound")
	cmd.Flags().StringVar(&opts.CreatedBefore, "created-before", "", "Creation date upper bound")
	cmd.Flags().StringVar(&opts.UpdatedAfter, "updated-after", "", "Update date lower bound")
	cmd.Flags().StringVar(&opts.UpdatedBefore, "updated-before", "", "Update date upper bound")
	cmd.Flags().StringVar(&opts.Creator, "creator", "", "Comma-separated creator IDs")
	cmd.Flags().StringVar(&opts.AssignedTo, "assigned-to", "", "Comma-separated assignee IDs")
	cmd.Flags().StringVar(&opts.Sprint, "sprint", "", "Comma-separated sprint IDs")
	cmd.Flags().StringVar(&opts.WorkitemType, "workitem-type", "", "Comma-separated workitem type IDs")
	cmd.Flags().StringVar(&opts.StatusStage, "status-stage", "", "Comma-separated status stages")
	cmd.Flags().StringVar(&opts.Tag, "tag", "", "Comma-separated tag IDs")
	cmd.Flags().StringVar(&opts.Priority, "priority", "", "Comma-separated priority IDs")
	cmd.Flags().StringVar(&opts.SubjectDescription, "subject-description", "", "Subject or description keyword filter")
	cmd.Flags().StringVar(&opts.FinishTimeAfter, "finish-time-after", "", "Finish date lower bound")
	cmd.Flags().StringVar(&opts.FinishTimeBefore, "finish-time-before", "", "Finish date upper bound")
	cmd.Flags().StringVar(&opts.UpdateStatusAtAfter, "update-status-at-after", "", "Status update date lower bound")
	cmd.Flags().StringVar(&opts.UpdateStatusAtBefore, "update-status-at-before", "", "Status update date upper bound")
	cmd.Flags().StringVar(&opts.AdvancedConditions, "advanced-conditions", "", "Raw Projex conditions JSON")
	cmd.Flags().StringVar(&opts.OrderBy, "order-by", "", "Workitem order field")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", "Sort direction")
	cmd.Flags().BoolVar(&mine, "mine", false, "List workitems assigned to the current user across participated projects")
	cmd.Flags().BoolVar(&unfinished, "unfinished", false, "Filter out completed workitems from returned items")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size; upstream fetch size when mine is set")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	flagmeta.MustMarkRequired(cmd, "category")
	return cmd
}

func listMyWorkitems(ctx context.Context, client *httpx.Client, organizationID, category string, pageSize int, opts projexdomain.WorkitemListOptions, unfinished bool) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	currentUser, errDetail := orgdomain.GetCurrentUser(ctx, client)
	if errDetail != nil {
		return nil, nil, errDetail
	}
	if currentUser.UserID == "" {
		return nil, nil, &output.ErrorDetail{Code: "CURRENT_USER_UNRESOLVED", Category: "general", Retryable: false, Message: "failed to resolve current user id from org current response"}
	}
	orgID := config.GetOrganizationID(organizationID)
	if orgID == "" && !config.IsRegionBaseURL(client.BaseURL) {
		if currentUser.LastOrganizationID == "" {
			return nil, nil, &output.ErrorDetail{Code: "PARAM_REQUIRED", Category: "param", Retryable: false, Message: "organization_id is required because current user has no lastOrganization"}
		}
		orgID = currentUser.LastOrganizationID
	}

	projectOpts := projexdomain.ProjectListOptions{ScenarioFilter: "participate", UserID: currentUser.UserID}
	projects, errDetail := listAllProjects(ctx, client, orgID, pageSize, projectOpts)
	if errDetail != nil {
		return nil, nil, addErrorContext(errDetail, "failed to list participated projects")
	}

	items := make([]map[string]any, 0)
	workitemOpts := opts
	workitemOpts.AssignedTo = currentUser.UserID
	for i, project := range projects {
		projectID, ok := strictStringMapField(project, "id")
		if !ok {
			return nil, nil, &output.ErrorDetail{Code: "RESPONSE_DECODE_FAILED", Category: "general", Retryable: false, Message: fmt.Sprintf("project list response item %d has no string id; cannot aggregate personal workitems", i)}
		}
		data, errDetail := listAllWorkitems(ctx, client, orgID, category, projectID, pageSize, workitemOpts)
		if errDetail != nil {
			return nil, nil, addErrorContext(errDetail, fmt.Sprintf("failed to list assigned workitems for project %s", projectID))
		}
		items = append(items, data...)
	}
	if unfinished {
		items, errDetail = filterUnfinishedWorkitems(items)
		if errDetail != nil {
			return nil, nil, errDetail
		}
	}
	total := len(items)
	return items, &output.Pagination{NextToken: nil, PageSize: pageSize, HasMore: false, Total: &total}, nil
}

func listAllProjects(ctx context.Context, client *httpx.Client, organizationID string, pageSize int, opts projexdomain.ProjectListOptions) ([]map[string]any, *output.ErrorDetail) {
	items := make([]map[string]any, 0)
	pageToken := ""
	seenTokens := map[string]bool{}
	contextLabel := "participated project search"
	for range maxAggregatePages {
		data, pagination, errDetail := projexdomain.ListProjects(ctx, client, organizationID, pageSize, pageToken, opts)
		if errDetail != nil {
			return nil, errDetail
		}
		items = append(items, data...)
		nextToken, hasMore, errDetail := nextAggregatePageToken(pagination, len(items), seenTokens, contextLabel)
		if errDetail != nil {
			return nil, errDetail
		}
		if !hasMore {
			return items, nil
		}
		seenTokens[nextToken] = true
		pageToken = nextToken
	}
	return nil, aggregatePageLimitError(contextLabel)
}

func listAllWorkitems(ctx context.Context, client *httpx.Client, organizationID, category, projectID string, pageSize int, opts projexdomain.WorkitemListOptions) ([]map[string]any, *output.ErrorDetail) {
	items := make([]map[string]any, 0)
	pageToken := ""
	seenTokens := map[string]bool{}
	contextLabel := fmt.Sprintf("workitem search for project %s", projectID)
	for range maxAggregatePages {
		data, pagination, errDetail := projexdomain.ListWorkitems(ctx, client, organizationID, category, projectID, pageSize, pageToken, opts)
		if errDetail != nil {
			return nil, errDetail
		}
		items = append(items, data...)
		nextToken, hasMore, errDetail := nextAggregatePageToken(pagination, len(items), seenTokens, contextLabel)
		if errDetail != nil {
			return nil, errDetail
		}
		if !hasMore {
			return items, nil
		}
		seenTokens[nextToken] = true
		pageToken = nextToken
	}
	return nil, aggregatePageLimitError(contextLabel)
}

func aggregatePageLimitError(contextLabel string) *output.ErrorDetail {
	return &output.ErrorDetail{Code: "PAGINATION_INVALID", Category: "general", Retryable: false, Message: fmt.Sprintf("%s exceeded %d pages without completing", contextLabel, maxAggregatePages)}
}

func nextAggregatePageToken(pagination *output.Pagination, itemCount int, seenTokens map[string]bool, contextLabel string) (string, bool, *output.ErrorDetail) {
	if pagination == nil {
		return "", false, &output.ErrorDetail{Code: "PAGINATION_INVALID", Category: "general", Retryable: false, Message: fmt.Sprintf("%s returned no pagination metadata after collecting %d items", contextLabel, itemCount)}
	}
	if !pagination.HasMore {
		if pagination.Total != nil && itemCount < *pagination.Total {
			return "", false, &output.ErrorDetail{Code: "PAGINATION_INVALID", Category: "general", Retryable: false, Message: fmt.Sprintf("%s returned %d of %d items but did not provide a next page token", contextLabel, itemCount, *pagination.Total)}
		}
		return "", false, nil
	}
	if pagination.NextToken == nil || strings.TrimSpace(*pagination.NextToken) == "" {
		return "", false, &output.ErrorDetail{Code: "PAGINATION_INVALID", Category: "general", Retryable: false, Message: fmt.Sprintf("%s reported more pages but did not provide a next page token", contextLabel)}
	}
	nextToken := strings.TrimSpace(*pagination.NextToken)
	if seenTokens[nextToken] {
		return "", false, &output.ErrorDetail{Code: "PAGINATION_LOOP_DETECTED", Category: "general", Retryable: false, Message: fmt.Sprintf("%s returned repeated next page token %q", contextLabel, nextToken)}
	}
	return nextToken, true, nil
}

func filterUnfinishedWorkitems(items []map[string]any) ([]map[string]any, *output.ErrorDetail) {
	filtered := make([]map[string]any, 0, len(items))
	for i, item := range items {
		completed, known := classifyWorkitemCompletion(item)
		if !known {
			workitemID, _ := strictStringMapField(item, "id")
			statusHint := workitemStatusHint(item)
			message := fmt.Sprintf("cannot apply --unfinished because aggregate item %d has no recognizable completion status: %s", i, statusHint)
			if workitemID != "" {
				message = fmt.Sprintf("cannot apply --unfinished because workitem %s has no recognizable completion status: %s", workitemID, statusHint)
			}
			return nil, &output.ErrorDetail{Code: "WORKITEM_STATUS_UNCLASSIFIED", Category: "general", Retryable: false, Message: message}
		}
		if !completed {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

type workitemCompletionStatus int

const (
	workitemCompletionUnknown workitemCompletionStatus = iota
	workitemCompletionIncomplete
	workitemCompletionCompleted
	workitemCompletionConflict
)

func classifyWorkitemCompletion(item map[string]any) (bool, bool) {
	logicalStatus := classifyStatusTexts(statusTextValues(item["logicalStatus"]))
	status := classifyStatusTexts(statusTextValues(item["status"]))
	if logicalStatus == workitemCompletionConflict || status == workitemCompletionConflict {
		return false, false
	}
	if logicalStatus != workitemCompletionUnknown && status != workitemCompletionUnknown && logicalStatus != status {
		return false, false
	}
	if logicalStatus != workitemCompletionUnknown {
		return logicalStatus == workitemCompletionCompleted, true
	}
	if status != workitemCompletionUnknown {
		return status == workitemCompletionCompleted, true
	}
	return false, false
}

func classifyStatusTexts(values []string) workitemCompletionStatus {
	hasIncomplete := slices.ContainsFunc(values, isIncompleteStatusText)
	hasCompleted := slices.ContainsFunc(values, isCompletedStatusText)
	switch {
	case hasIncomplete && hasCompleted:
		return workitemCompletionConflict
	case hasIncomplete:
		return workitemCompletionIncomplete
	case hasCompleted:
		return workitemCompletionCompleted
	default:
		return workitemCompletionUnknown
	}
}

func workitemStatusHint(item map[string]any) string {
	values := append(statusTextValues(item["logicalStatus"]), statusTextValues(item["status"])...)
	if len(values) == 0 {
		return "logicalStatus/status are empty"
	}
	return "logicalStatus/status values: " + strings.Join(values, ", ")
}

func statusTextValues(value any) []string {
	switch typed := value.(type) {
	case string:
		return []string{typed}
	case map[string]any:
		keys := []string{"name", "displayName", "label", "state", "category", "stage", "logicalStatus"}
		values := make([]string, 0, len(keys))
		for _, key := range keys {
			if text, ok := typed[key].(string); ok && text != "" {
				values = append(values, text)
			}
		}
		return values
	default:
		return nil
	}
}

func isIncompleteStatusText(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "open", "todo", "to_do", "pending", "in_progress", "in progress", "in-progress", "doing", "待处理", "待办", "处理中", "进行中", "打开", "未开始", "未完成":
		return true
	default:
		return strings.Contains(normalized, "未完成")
	}
}

func isCompletedStatusText(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "done", "completed", "closed", "resolved", "完成", "关闭", "解决", "已完成", "已关闭", "已解决":
		return true
	default:
		return strings.Contains(normalized, "已完成") || strings.Contains(normalized, "已关闭") || strings.Contains(normalized, "已解决")
	}
}

func addErrorContext(errDetail *output.ErrorDetail, context string) *output.ErrorDetail {
	if errDetail == nil {
		return nil
	}
	cloned := *errDetail
	cloned.Message = context + ": " + errDetail.Message
	return &cloned
}

func strictStringMapField(data map[string]any, key string) (string, bool) {
	value, ok := data[key]
	if !ok || value == nil {
		return "", false
	}
	str, ok := value.(string)
	if !ok || str == "" {
		return "", false
	}
	return str, true
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
