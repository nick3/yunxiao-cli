package flow

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/aliyun/yunxiao-cli/internal/domains/shared"
	"github.com/aliyun/yunxiao-cli/internal/httpx"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
)

type ListPipelinesResponse struct {
	Data     []map[string]any `json:"data"`
	NextPage any              `json:"nextPage"`
}

func ListPipelines(ctx context.Context, client *httpx.Client, organizationID string, pageSize int, pageToken string) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	var apiResp ListPipelinesResponse
	path := pipelinesPath(client.BaseURL, organizationID)
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

func GetPipeline(ctx context.Context, client *httpx.Client, organizationID, pipelineID string) (map[string]any, *output.ErrorDetail) {
	var data map[string]any
	path := pipelinesPath(client.BaseURL, organizationID) + "/" + url.PathEscape(pipelineID)
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path, &data); errDetail != nil {
		return nil, errDetail
	}
	return data, nil
}

func pipelinesPath(baseURL, organizationID string) string {
	if shared.IsRegionBaseURL(baseURL) {
		return "/oapi/v1/flow/pipelines"
	}
	return "/oapi/v1/flow/organizations/" + url.PathEscape(organizationID) + "/pipelines"
}
