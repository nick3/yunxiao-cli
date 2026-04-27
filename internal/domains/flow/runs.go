package flow

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

type RunListOptions struct {
	StartTime   string
	EndTime     string
	Status      string
	TriggerMode string
}

func ListRuns(ctx context.Context, client *httpx.Client, organizationID, pipelineID string, pageSize int, pageToken string, opts RunListOptions) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	path := pipelinesPath(client.BaseURL, organizationID) + "/" + url.PathEscape(pipelineID) + "/runs"
	query := url.Values{"perPage": []string{strconv.Itoa(pageSize)}}
	if pageToken != "" {
		query.Set("page", pageToken)
	}
	setQuery(query, "startTime", opts.StartTime)
	setQuery(query, "endTime", opts.EndTime)
	setQuery(query, "status", opts.Status)
	setQuery(query, "triggerMode", opts.TriggerMode)

	var body json.RawMessage
	headers, errDetail := shared.RequestJSONWithBodyAndHeaders(ctx, client, http.MethodGet, path+"?"+query.Encode(), nil, &body)
	if errDetail != nil {
		return nil, nil, errDetail
	}
	return shared.DecodeSearchList(body, headers, pageSize)
}

func GetRun(ctx context.Context, client *httpx.Client, organizationID, pipelineID, runID string) (map[string]any, *output.ErrorDetail) {
	var data map[string]any
	path := pipelinesPath(client.BaseURL, organizationID) + "/" + url.PathEscape(pipelineID) + "/runs/" + url.PathEscape(runID)
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path, &data); errDetail != nil {
		return nil, errDetail
	}
	return data, nil
}

func setQuery(query url.Values, key, value string) {
	if value != "" {
		query.Set(key, value)
	}
}
