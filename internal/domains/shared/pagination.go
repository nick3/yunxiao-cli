package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/aliyun/yunxiao-cli/internal/model/output"
)

// SearchResponse is the common envelope returned by Yunxiao `:search` style endpoints.
type SearchResponse struct {
	Data     []map[string]any `json:"data"`
	NextPage any              `json:"nextPage"`
}

func DecodeSearchList(body json.RawMessage, headers http.Header, pageSize int) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return nil, nil, &output.ErrorDetail{Code: "EMPTY_RESPONSE", Category: "general", Retryable: false, Message: "upstream returned an empty response body for a search response"}
	}

	if body[0] == '[' {
		var data []map[string]any
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, nil, decodeSearchError(err)
		}
		pagination, errDetail := SearchPaginationFromHeadersStrict(headers, pageSize)
		if errDetail != nil {
			return nil, nil, errDetail
		}
		return data, pagination, nil
	}

	var apiResp SearchResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, nil, decodeSearchError(err)
	}
	nextToken := StringToken(apiResp.NextPage)
	pagination, errDetail := SearchPaginationFromHeadersStrict(headers, pageSize)
	if errDetail != nil {
		return nil, nil, errDetail
	}
	pagination.NextToken = nextToken
	pagination.HasMore = nextToken != nil
	return apiResp.Data, pagination, nil
}

func SearchPaginationFromHeaders(headers http.Header, fallbackPageSize int) *output.Pagination {
	nextToken := StringToken(headers.Get("x-next-page"))
	pageSize := fallbackPageSize
	if perPage, err := strconv.Atoi(strings.TrimSpace(headers.Get("x-per-page"))); err == nil && perPage > 0 {
		pageSize = perPage
	}
	return &output.Pagination{
		NextToken:  nextToken,
		PageSize:   pageSize,
		HasMore:    nextToken != nil,
		Page:       headerInt(headers, "x-page"),
		TotalPages: headerInt(headers, "x-total-pages"),
		Total:      headerInt(headers, "x-total"),
		PrevToken:  StringToken(headers.Get("x-prev-page")),
	}
}

func SearchPaginationFromHeadersStrict(headers http.Header, fallbackPageSize int) (*output.Pagination, *output.ErrorDetail) {
	page, errDetail := strictHeaderInt(headers, "x-page")
	if errDetail != nil {
		return nil, errDetail
	}
	totalPages, errDetail := strictHeaderInt(headers, "x-total-pages")
	if errDetail != nil {
		return nil, errDetail
	}
	total, errDetail := strictHeaderInt(headers, "x-total")
	if errDetail != nil {
		return nil, errDetail
	}
	pagination := SearchPaginationFromHeaders(headers, fallbackPageSize)
	pagination.Page = page
	pagination.TotalPages = totalPages
	pagination.Total = total
	return pagination, nil
}

func strictHeaderInt(headers http.Header, key string) (*int, *output.ErrorDetail) {
	raw := strings.TrimSpace(headers.Get(key))
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return nil, &output.ErrorDetail{Code: "PAGINATION_INVALID", Category: "general", Retryable: false, Message: fmt.Sprintf("upstream returned invalid pagination header %s=%q", key, raw)}
	}
	return &value, nil
}

func headerInt(headers http.Header, key string) *int {
	raw := strings.TrimSpace(headers.Get(key))
	if raw == "" {
		return nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return nil
	}
	return &value
}

// ApplyPageToken sets the "page" key on a search payload, preferring an int when the token parses cleanly.
func ApplyPageToken(payload map[string]any, pageToken string) {
	if pageToken == "" {
		return
	}
	if page, err := strconv.Atoi(pageToken); err == nil {
		payload["page"] = page
		return
	}
	payload["page"] = pageToken
}

func decodeSearchError(err error) *output.ErrorDetail {
	return &output.ErrorDetail{Code: "RESPONSE_DECODE_FAILED", Category: "general", Retryable: false, Message: fmt.Sprintf("failed to decode search response: %v", err)}
}
