package projex

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/aliyun/yunxiao-cli/internal/domains/shared"
	"github.com/aliyun/yunxiao-cli/internal/httpx"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
)

func ListProjects(ctx context.Context, client *httpx.Client, organizationID string, pageSize int, pageToken string) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	payload := map[string]any{"perPage": pageSize}
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

func ListWorkitems(ctx context.Context, client *httpx.Client, organizationID, category, spaceID string, pageSize int, pageToken string) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	payload := map[string]any{"category": category, "spaceId": spaceID, "perPage": pageSize}
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
	nextToken := shared.StringToken(headers.Get("x-next-page"))
	return data, &output.Pagination{NextToken: nextToken, PageSize: pageSize, HasMore: nextToken != nil}, nil
}

func searchList(ctx context.Context, client *httpx.Client, path string, payload map[string]any, pageSize int) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	var apiResp shared.SearchResponse
	if errDetail := shared.RequestJSONWithBody(ctx, client, http.MethodPost, path, payload, &apiResp); errDetail != nil {
		return nil, nil, errDetail
	}
	nextToken := shared.StringToken(apiResp.NextPage)
	return apiResp.Data, &output.Pagination{NextToken: nextToken, PageSize: pageSize, HasMore: nextToken != nil}, nil
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
