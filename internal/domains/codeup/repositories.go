package codeup

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/aliyun/yunxiao-cli/internal/domains/shared"
	"github.com/aliyun/yunxiao-cli/internal/httpx"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
)

type ListRepositoriesResponse struct {
	Data     []map[string]any `json:"data"`
	NextPage any              `json:"nextPage"`
}

func ListRepositories(ctx context.Context, client *httpx.Client, organizationID string, pageSize int, pageToken string) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	var apiResp ListRepositoriesResponse
	path := repositoriesPath(client.BaseURL, organizationID)
	query := url.Values{}
	query.Set("perPage", strconv.Itoa(pageSize))
	if pageToken != "" {
		query.Set("page", pageToken)
	}
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path, &apiResp); errDetail != nil {
		return nil, nil, errDetail
	}
	nextToken := shared.StringToken(apiResp.NextPage)
	pagination := &output.Pagination{NextToken: nextToken, PageSize: pageSize, HasMore: nextToken != nil}
	return apiResp.Data, pagination, nil
}

func GetRepository(ctx context.Context, client *httpx.Client, organizationID, repoID string) (map[string]any, *output.ErrorDetail) {
	var data map[string]any
	path := repositoriesPath(client.BaseURL, organizationID) + "/" + url.PathEscape(repoID)
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path, &data); errDetail != nil {
		return nil, errDetail
	}
	return data, nil
}

func repositoriesPath(baseURL, organizationID string) string {
	if shared.IsRegionBaseURL(baseURL) {
		return "/oapi/v1/codeup/repositories"
	}
	return "/oapi/v1/codeup/organizations/" + url.PathEscape(organizationID) + "/repositories"
}
