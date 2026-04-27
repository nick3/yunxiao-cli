package projex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/nick3/yunxiao-cli/internal/domains/shared"
	"github.com/nick3/yunxiao-cli/internal/httpx"
	"github.com/nick3/yunxiao-cli/internal/model/output"
)

func ListWorkitemComments(ctx context.Context, client *httpx.Client, organizationID, workitemID string, pageSize int, pageToken string) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	query := url.Values{"perPage": []string{strconv.Itoa(pageSize)}}
	if pageToken != "" {
		query.Set("page", pageToken)
	}
	path := workitemsPath(client.BaseURL, organizationID) + "/" + url.PathEscape(workitemID) + "/comments?" + query.Encode()

	var body json.RawMessage
	headers, errDetail := shared.RequestJSONWithBodyAndHeaders(ctx, client, http.MethodGet, path, nil, &body)
	if errDetail != nil {
		return nil, nil, errDetail
	}
	data, errDetail := decodeArrayOrResult(body)
	if errDetail != nil {
		return nil, nil, errDetail
	}
	pagination, errDetail := shared.SearchPaginationFromHeadersStrict(headers, pageSize)
	if errDetail != nil {
		return nil, nil, errDetail
	}
	return data, pagination, nil
}

func CreateWorkitemComment(ctx context.Context, client *httpx.Client, organizationID, workitemID, content string) (map[string]any, *output.ErrorDetail) {
	var body json.RawMessage
	path := workitemsPath(client.BaseURL, organizationID) + "/" + url.PathEscape(workitemID) + "/comments"
	if errDetail := shared.RequestJSONWithBody(ctx, client, http.MethodPost, path, map[string]any{"content": content}, &body); errDetail != nil {
		return nil, errDetail
	}
	return decodeResourceObjectOrResult(body, "workitem comment create", "id", "identifier", "commentId", "commentID")
}
