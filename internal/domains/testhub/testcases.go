package testhub

import (
	"context"
	"net/http"
	"net/url"

	"github.com/nick3/yunxiao-cli/internal/domains/shared"
	"github.com/nick3/yunxiao-cli/internal/httpx"
	"github.com/nick3/yunxiao-cli/internal/model/output"
)

func ListTestcases(ctx context.Context, client *httpx.Client, organizationID, testRepoID string, pageSize int, pageToken string) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	payload := map[string]any{"perPage": pageSize}
	shared.ApplyPageToken(payload, pageToken)
	var apiResp shared.SearchResponse
	if errDetail := shared.RequestJSONWithBody(ctx, client, http.MethodPost, testcasesPath(client.BaseURL, organizationID, testRepoID)+":search", payload, &apiResp); errDetail != nil {
		return nil, nil, errDetail
	}
	nextToken := shared.StringToken(apiResp.NextPage)
	return apiResp.Data, &output.Pagination{NextToken: nextToken, PageSize: pageSize, HasMore: nextToken != nil}, nil
}

func GetTestcase(ctx context.Context, client *httpx.Client, organizationID, testRepoID, testcaseID string) (map[string]any, *output.ErrorDetail) {
	var data map[string]any
	path := testcasesPath(client.BaseURL, organizationID, testRepoID) + "/" + url.PathEscape(testcaseID)
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path, &data); errDetail != nil {
		return nil, errDetail
	}
	return data, nil
}

func ListDirectories(ctx context.Context, client *httpx.Client, organizationID, testRepoID string) (map[string]any, *output.ErrorDetail) {
	var data map[string]any
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, directoriesPath(client.BaseURL, organizationID, testRepoID), &data); errDetail != nil {
		return nil, errDetail
	}
	return data, nil
}

func ListTestplans(ctx context.Context, client *httpx.Client, organizationID string) ([]map[string]any, *output.ErrorDetail) {
	var data []map[string]any
	if errDetail := shared.RequestJSONWithBody(ctx, client, http.MethodPost, testplansPath(client.BaseURL, organizationID), map[string]any{}, &data); errDetail != nil {
		return nil, errDetail
	}
	return data, nil
}

func testcasesPath(baseURL, organizationID, testRepoID string) string {
	if shared.IsRegionBaseURL(baseURL) {
		return "/oapi/v1/testhub/testRepos/" + url.PathEscape(testRepoID) + "/testcases"
	}
	return "/oapi/v1/testhub/organizations/" + url.PathEscape(organizationID) + "/testRepos/" + url.PathEscape(testRepoID) + "/testcases"
}

func directoriesPath(baseURL, organizationID, testRepoID string) string {
	if shared.IsRegionBaseURL(baseURL) {
		return "/oapi/v1/testhub/testRepos/" + url.PathEscape(testRepoID) + "/directories"
	}
	return "/oapi/v1/testhub/organizations/" + url.PathEscape(organizationID) + "/testRepos/" + url.PathEscape(testRepoID) + "/directories"
}

func testplansPath(baseURL, organizationID string) string {
	if shared.IsRegionBaseURL(baseURL) {
		return "/oapi/v1/projex/testPlan/list"
	}
	return "/oapi/v1/projex/organizations/" + url.PathEscape(organizationID) + "/testPlan/list"
}
