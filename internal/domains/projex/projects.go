package projex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/aliyun/yunxiao-cli/internal/domains/shared"
	"github.com/aliyun/yunxiao-cli/internal/httpx"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
)

func ListProjects(ctx context.Context, client *httpx.Client, organizationID string, pageSize int, pageToken string, opts ProjectListOptions) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	payload := map[string]any{"perPage": pageSize}
	if conditions := buildProjectConditions(opts); conditions != "" {
		payload["conditions"] = conditions
	}
	if extraConditions := buildProjectExtraConditions(opts); extraConditions != "" {
		payload["extraConditions"] = extraConditions
	}
	if opts.OrderBy != "" {
		payload["orderBy"] = opts.OrderBy
	}
	if opts.Sort != "" {
		payload["sort"] = opts.Sort
	}
	shared.ApplyPageToken(payload, pageToken)
	return searchList(ctx, client, projectsPath(client.BaseURL, organizationID)+":search", payload, pageSize)
}

func GetProject(ctx context.Context, client *httpx.Client, organizationID, projectID string) (map[string]any, *output.ErrorDetail) {
	var data map[string]any
	path := projectsPath(client.BaseURL, organizationID) + "/" + url.PathEscape(projectID)
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path, &data); errDetail != nil {
		return nil, errDetail
	}
	return data, nil
}

type ProjectListOptions struct {
	Name               string
	Status             string
	CreatedAfter       string
	CreatedBefore      string
	Creator            string
	AdminUserID        string
	LogicalStatus      string
	AdvancedConditions string
	ExtraConditions    string
	OrderBy            string
	Sort               string
	ScenarioFilter     string
	UserID             string
}

type WorkitemListOptions struct {
	SpaceType            string
	Subject              string
	Status               string
	CreatedAfter         string
	CreatedBefore        string
	UpdatedAfter         string
	UpdatedBefore        string
	Creator              string
	AssignedTo           string
	Sprint               string
	WorkitemType         string
	StatusStage          string
	Tag                  string
	Priority             string
	SubjectDescription   string
	FinishTimeAfter      string
	FinishTimeBefore     string
	UpdateStatusAtAfter  string
	UpdateStatusAtBefore string
	AdvancedConditions   string
	OrderBy              string
	Sort                 string
}

func ListWorkitems(ctx context.Context, client *httpx.Client, organizationID, category, spaceID string, pageSize int, pageToken string, opts WorkitemListOptions) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	payload := map[string]any{"category": category, "spaceId": spaceID, "perPage": pageSize}
	if opts.SpaceType != "" {
		payload["spaceType"] = opts.SpaceType
	}
	if conditions := buildWorkitemConditions(opts); conditions != "" {
		payload["conditions"] = conditions
	}
	if opts.OrderBy != "" {
		payload["orderBy"] = opts.OrderBy
	}
	if opts.Sort != "" {
		payload["sort"] = opts.Sort
	}
	shared.ApplyPageToken(payload, pageToken)
	return searchList(ctx, client, workitemsPath(client.BaseURL, organizationID)+":search", payload, pageSize)
}

func GetWorkitem(ctx context.Context, client *httpx.Client, organizationID, workitemID string) (map[string]any, *output.ErrorDetail) {
	var data map[string]any
	path := workitemsPath(client.BaseURL, organizationID) + "/" + url.PathEscape(workitemID)
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path, &data); errDetail != nil {
		return nil, errDetail
	}
	return data, nil
}

func ListSprints(ctx context.Context, client *httpx.Client, organizationID, projectID string, pageSize int, pageToken string, status string) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	var data []map[string]any
	query := url.Values{"perPage": []string{strconv.Itoa(pageSize)}}
	if pageToken != "" {
		query.Set("page", pageToken)
	}
	if status != "" {
		query.Set("status", status)
	}
	path := sprintsPath(client.BaseURL, organizationID, projectID) + "?" + query.Encode()
	headers, errDetail := shared.RequestJSONWithBodyAndHeaders(ctx, client, http.MethodGet, path, nil, &data)
	if errDetail != nil {
		return nil, nil, errDetail
	}
	return data, shared.SearchPaginationFromHeaders(headers, pageSize), nil
}

func searchList(ctx context.Context, client *httpx.Client, path string, payload map[string]any, pageSize int) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	var body json.RawMessage
	headers, errDetail := shared.RequestJSONWithBodyAndHeaders(ctx, client, http.MethodPost, path, payload, &body)
	if errDetail != nil {
		return nil, nil, errDetail
	}
	return shared.DecodeSearchList(body, headers, pageSize)
}

func buildProjectConditions(opts ProjectListOptions) string {
	if opts.AdvancedConditions != "" {
		return opts.AdvancedConditions
	}
	conditions := make([]map[string]any, 0)
	addStringCondition(&conditions, "string", "name", opts.Name)
	addListCondition(&conditions, "status", "status", "list", opts.Status)
	addDateCondition(&conditions, "date", "gmtCreate", opts.CreatedAfter, opts.CreatedBefore)
	addListCondition(&conditions, "user", "creator", "list", opts.Creator)
	addListCondition(&conditions, "user", "project.admin", "multiList", opts.AdminUserID)
	addListCondition(&conditions, "string", "logicalStatus", "list", opts.LogicalStatus)
	return marshalConditionGroups(conditions)
}

func buildProjectExtraConditions(opts ProjectListOptions) string {
	if opts.ScenarioFilter == "" || opts.UserID == "" {
		return opts.ExtraConditions
	}
	fieldIdentifier := ""
	switch opts.ScenarioFilter {
	case "manage":
		fieldIdentifier = "project.admin"
	case "participate":
		fieldIdentifier = "users"
	case "favorite":
		fieldIdentifier = "collectMembers"
	default:
		return opts.ExtraConditions
	}
	body, err := json.Marshal(map[string]any{"conditionGroups": []any{[]map[string]any{{"className": "user", "fieldIdentifier": fieldIdentifier, "format": "multiList", "operator": "CONTAINS", "value": []string{opts.UserID}}}}})
	if err != nil {
		return opts.ExtraConditions
	}
	return string(body)
}

func buildWorkitemConditions(opts WorkitemListOptions) string {
	if opts.AdvancedConditions != "" {
		return opts.AdvancedConditions
	}
	conditions := make([]map[string]any, 0)
	addStringCondition(&conditions, "string", "subject", opts.Subject)
	addListCondition(&conditions, "status", "status", "list", opts.Status)
	addDateCondition(&conditions, "dateTime", "gmtCreate", opts.CreatedAfter, opts.CreatedBefore)
	addDateCondition(&conditions, "dateTime", "gmtModified", opts.UpdatedAfter, opts.UpdatedBefore)
	addListCondition(&conditions, "user", "creator", "list", opts.Creator)
	addListCondition(&conditions, "user", "assignedTo", "list", opts.AssignedTo)
	addListCondition(&conditions, "sprint", "sprint", "list", opts.Sprint)
	addListCondition(&conditions, "workitemType", "workitemType", "list", opts.WorkitemType)
	addListCondition(&conditions, "statusStage", "statusStage", "list", opts.StatusStage)
	addListCondition(&conditions, "tag", "tag", "multiList", opts.Tag)
	addListCondition(&conditions, "option", "priority", "list", opts.Priority)
	addStringCondition(&conditions, "string", "subject-description", opts.SubjectDescription)
	addDateCondition(&conditions, "date", "finishTime", opts.FinishTimeAfter, opts.FinishTimeBefore)
	addDateCondition(&conditions, "date", "updateStatusAt", opts.UpdateStatusAtAfter, opts.UpdateStatusAtBefore)
	return marshalConditionGroups(conditions)
}

func marshalConditionGroups(conditions []map[string]any) string {
	if len(conditions) == 0 {
		return ""
	}
	body, err := json.Marshal(map[string]any{"conditionGroups": []any{conditions}})
	if err != nil {
		return ""
	}
	return string(body)
}

func addStringCondition(conditions *[]map[string]any, className, fieldIdentifier, value string) {
	if value == "" {
		return
	}
	*conditions = append(*conditions, map[string]any{"className": className, "fieldIdentifier": fieldIdentifier, "format": "input", "operator": "CONTAINS", "toValue": nil, "value": []string{value}})
}

func addListCondition(conditions *[]map[string]any, className, fieldIdentifier, format, value string) {
	if value == "" {
		return
	}
	*conditions = append(*conditions, map[string]any{"className": className, "fieldIdentifier": fieldIdentifier, "format": format, "operator": "CONTAINS", "toValue": nil, "value": splitCSV(value)})
}

func addDateCondition(conditions *[]map[string]any, className, fieldIdentifier, after, before string) {
	if after == "" && before == "" {
		return
	}
	var value any
	if after != "" {
		value = []string{after + " 00:00:00"}
	}
	var toValue any
	if before != "" {
		toValue = before + " 23:59:59"
	}
	*conditions = append(*conditions, map[string]any{"className": className, "fieldIdentifier": fieldIdentifier, "format": "input", "operator": "BETWEEN", "toValue": toValue, "value": value})
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func projectsPath(baseURL, organizationID string) string {
	if shared.IsRegionBaseURL(baseURL) {
		return "/oapi/v1/projex/projects"
	}
	return "/oapi/v1/projex/organizations/" + url.PathEscape(organizationID) + "/projects"
}

func workitemsPath(baseURL, organizationID string) string {
	if shared.IsRegionBaseURL(baseURL) {
		return "/oapi/v1/projex/workitems"
	}
	return "/oapi/v1/projex/organizations/" + url.PathEscape(organizationID) + "/workitems"
}

func sprintsPath(baseURL, organizationID, projectID string) string {
	if shared.IsRegionBaseURL(baseURL) {
		return "/oapi/v1/projex/projects/" + url.PathEscape(projectID) + "/sprints"
	}
	return "/oapi/v1/projex/organizations/" + url.PathEscape(organizationID) + "/projects/" + url.PathEscape(projectID) + "/sprints"
}
