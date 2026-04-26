package org

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/aliyun/yunxiao-cli/internal/domains/shared"
	"github.com/aliyun/yunxiao-cli/internal/httpx"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
)

func ListMembers(ctx context.Context, client *httpx.Client, organizationID string, pageSize int, pageToken string) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	path := membersPath(client.BaseURL, organizationID)
	query := url.Values{"perPage": []string{strconv.Itoa(pageSize)}}
	if pageToken != "" {
		query.Set("page", pageToken)
	}

	var body json.RawMessage
	headers, errDetail := shared.RequestJSONWithBodyAndHeaders(ctx, client, http.MethodGet, path+"?"+query.Encode(), nil, &body)
	if errDetail != nil {
		return nil, nil, errDetail
	}
	return shared.DecodeSearchList(body, headers, pageSize)
}

func membersPath(baseURL, organizationID string) string {
	if shared.IsRegionBaseURL(baseURL) {
		return "/oapi/v1/platform/members"
	}
	return "/oapi/v1/platform/organizations/" + url.PathEscape(organizationID) + "/members"
}
