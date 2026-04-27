package codeup

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

func ListBranches(ctx context.Context, client *httpx.Client, organizationID, repoID string, pageSize int, pageToken, sort, search string) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	path := repositoryPath(client.BaseURL, organizationID, repoID) + "/branches"
	query := url.Values{"perPage": []string{strconv.Itoa(pageSize)}}
	if pageToken != "" {
		query.Set("page", pageToken)
	}
	if sort != "" {
		query.Set("sort", sort)
	}
	if search != "" {
		query.Set("search", search)
	}
	return listRaw(ctx, client, path+"?"+query.Encode(), pageSize)
}

func ListCommits(ctx context.Context, client *httpx.Client, organizationID, repoID string, pageSize int, pageToken string, opts CommitListOptions) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	path := repositoryPath(client.BaseURL, organizationID, repoID) + "/commits"
	query := url.Values{"perPage": []string{strconv.Itoa(pageSize)}}
	if pageToken != "" {
		query.Set("page", pageToken)
	}
	setQuery(query, "refName", opts.RefName)
	setQuery(query, "since", opts.Since)
	setQuery(query, "until", opts.Until)
	setQuery(query, "path", opts.Path)
	setQuery(query, "search", opts.Search)
	setQuery(query, "committerIds", opts.CommitterIDs)
	if opts.ShowSignature != nil {
		query.Set("showSignature", strconv.FormatBool(*opts.ShowSignature))
	}
	return listRaw(ctx, client, path+"?"+query.Encode(), pageSize)
}

func GetFile(ctx context.Context, client *httpx.Client, organizationID, repoID, filePath, ref string) (map[string]any, *output.ErrorDetail) {
	var data map[string]any
	path := repositoryPath(client.BaseURL, organizationID, repoID) + "/files/" + url.PathEscape(filePath)
	query := url.Values{"ref": []string{ref}}
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path+"?"+query.Encode(), &data); errDetail != nil {
		return nil, errDetail
	}
	return data, nil
}

func GetCompare(ctx context.Context, client *httpx.Client, organizationID, repoID string, opts CompareOptions) (map[string]any, *output.ErrorDetail) {
	var data map[string]any
	path := repositoryPath(client.BaseURL, organizationID, repoID) + "/compares"
	query := url.Values{"from": []string{opts.From}, "to": []string{opts.To}}
	setQuery(query, "sourceType", opts.SourceType)
	setQuery(query, "targetType", opts.TargetType)
	if opts.Straight != nil {
		query.Set("straight", strconv.FormatBool(*opts.Straight))
	}
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path+"?"+query.Encode(), &data); errDetail != nil {
		return nil, errDetail
	}
	return data, nil
}

type CommitListOptions struct {
	RefName       string
	Since         string
	Until         string
	Path          string
	Search        string
	CommitterIDs  string
	ShowSignature *bool
}

type CompareOptions struct {
	From       string
	To         string
	SourceType string
	TargetType string
	Straight   *bool
}

func listRaw(ctx context.Context, client *httpx.Client, path string, pageSize int) ([]map[string]any, *output.Pagination, *output.ErrorDetail) {
	var body json.RawMessage
	headers, errDetail := shared.RequestJSONWithBodyAndHeaders(ctx, client, http.MethodGet, path, nil, &body)
	if errDetail != nil {
		return nil, nil, errDetail
	}
	return shared.DecodeSearchList(body, headers, pageSize)
}

func repositoryPath(baseURL, organizationID, repoID string) string {
	return repositoriesPath(baseURL, organizationID) + "/" + url.PathEscape(repoID)
}

func setQuery(query url.Values, key, value string) {
	if value != "" {
		query.Set(key, value)
	}
}
