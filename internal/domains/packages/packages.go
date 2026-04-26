package packages

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/aliyun/yunxiao-cli/internal/domains/shared"
	"github.com/aliyun/yunxiao-cli/internal/httpx"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
)

func ListRepositories(ctx context.Context, client *httpx.Client, organizationID string, pageSize int, pageToken string, repoTypes string, repoCategories string) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	var data []map[string]any
	path := repositoriesPath(client.BaseURL, organizationID)
	query := url.Values{}
	query.Set("perPage", strconv.Itoa(pageSize))
	if pageToken != "" {
		query.Set("page", pageToken)
	}
	if repoTypes != "" {
		query.Set("repoTypes", repoTypes)
	}
	if repoCategories != "" {
		query.Set("repoCategories", repoCategories)
	}
	path += "?" + query.Encode()
	headers, errDetail := shared.RequestJSONWithBodyAndHeaders(ctx, client, http.MethodGet, path, nil, &data)
	if errDetail != nil {
		return nil, nil, errDetail
	}
	return data, packagePagination(headers.Get("x-next-page"), pageSize), nil
}

func ListArtifacts(ctx context.Context, client *httpx.Client, organizationID, repoID, repoType string, pageSize int, pageToken string, search string) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	var data []map[string]any
	path := artifactsPath(client.BaseURL, organizationID, repoID)
	query := url.Values{"repoType": []string{repoType}, "perPage": []string{strconv.Itoa(pageSize)}, "orderBy": []string{"latestUpdate"}, "sort": []string{"desc"}}
	if pageToken != "" {
		query.Set("page", pageToken)
	}
	if search != "" {
		query.Set("search", search)
	}
	path += "?" + query.Encode()
	headers, errDetail := shared.RequestJSONWithBodyAndHeaders(ctx, client, http.MethodGet, path, nil, &data)
	if errDetail != nil {
		return nil, nil, errDetail
	}
	return data, packagePagination(headers.Get("x-next-page"), pageSize), nil
}

func packagePagination(nextPage string, pageSize int) *output.Pagination {
	if nextPage == "" {
		return &output.Pagination{NextToken: nil, PageSize: pageSize, HasMore: false}
	}
	return &output.Pagination{NextToken: &nextPage, PageSize: pageSize, HasMore: true}
}

func GetArtifact(ctx context.Context, client *httpx.Client, organizationID, repoID, artifactID, repoType string) (map[string]any, *output.ErrorDetail) {
	var data map[string]any
	path := artifactsPath(client.BaseURL, organizationID, repoID) + "/" + url.PathEscape(artifactID)
	query := url.Values{"repoType": []string{repoType}}
	path += "?" + query.Encode()
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path, &data); errDetail != nil {
		return nil, errDetail
	}
	return data, nil
}

func repositoriesPath(baseURL, organizationID string) string {
	if shared.IsRegionBaseURL(baseURL) {
		return "/oapi/v1/packages/repositories"
	}
	return "/oapi/v1/packages/organizations/" + url.PathEscape(organizationID) + "/repositories"
}

func artifactsPath(baseURL, organizationID, repoID string) string {
	if shared.IsRegionBaseURL(baseURL) {
		return "/oapi/v1/packages/repositories/" + url.PathEscape(repoID) + "/artifacts"
	}
	return "/oapi/v1/packages/organizations/" + url.PathEscape(organizationID) + "/repositories/" + url.PathEscape(repoID) + "/artifacts"
}
