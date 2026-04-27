package shared

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeSearchListRawArrayUsesHeaderPagination(t *testing.T) {
	headers := http.Header{}
	headers.Set("x-next-page", "2")
	headers.Set("x-per-page", "5")

	data, pagination, errDetail := DecodeSearchList(json.RawMessage(`[{"id":"proj-1"}]`), headers, 20)

	require.Nil(t, errDetail)
	require.Equal(t, []map[string]any{{"id": "proj-1"}}, data)
	require.NotNil(t, pagination.NextToken)
	require.Equal(t, "2", *pagination.NextToken)
	require.Equal(t, 5, pagination.PageSize)
	require.True(t, pagination.HasMore)
}

func TestDecodeSearchListRawArrayIncludesHeaderTotals(t *testing.T) {
	headers := http.Header{}
	headers.Set("x-page", "2")
	headers.Set("x-next-page", "3")
	headers.Set("x-prev-page", "1")
	headers.Set("x-per-page", "5")
	headers.Set("x-total-pages", "10")
	headers.Set("x-total", "47")

	_, pagination, errDetail := DecodeSearchList(json.RawMessage(`[{}]`), headers, 20)

	require.Nil(t, errDetail)
	require.NotNil(t, pagination.Page)
	require.Equal(t, 2, *pagination.Page)
	require.NotNil(t, pagination.NextToken)
	require.Equal(t, "3", *pagination.NextToken)
	require.NotNil(t, pagination.PrevToken)
	require.Equal(t, "1", *pagination.PrevToken)
	require.Equal(t, 5, pagination.PageSize)
	require.NotNil(t, pagination.TotalPages)
	require.Equal(t, 10, *pagination.TotalPages)
	require.NotNil(t, pagination.Total)
	require.Equal(t, 47, *pagination.Total)
	require.True(t, pagination.HasMore)
}

func TestDecodeSearchListRawArrayMissingNextPageHasNoContinuation(t *testing.T) {
	data, pagination, errDetail := DecodeSearchList(json.RawMessage(`[]`), http.Header{}, 20)

	require.Nil(t, errDetail)
	require.Empty(t, data)
	require.Nil(t, pagination.NextToken)
	require.Equal(t, 20, pagination.PageSize)
	require.False(t, pagination.HasMore)
}

func TestDecodeSearchListRejectsInvalidPerPageHeader(t *testing.T) {
	for _, value := range []string{"invalid", "0", "-1"} {
		t.Run(value, func(t *testing.T) {
			headers := http.Header{}
			headers.Set("x-next-page", "2")
			headers.Set("x-per-page", value)

			data, pagination, errDetail := DecodeSearchList(json.RawMessage(`[{"id":"proj-1"}]`), headers, 20)

			require.Nil(t, data)
			require.Nil(t, pagination)
			require.NotNil(t, errDetail)
			require.Equal(t, "PAGINATION_INVALID", errDetail.Code)
			require.Contains(t, errDetail.Message, "x-per-page")
			require.Contains(t, errDetail.Message, value)
		})
	}
}

func TestDecodeSearchListRejectsInvalidCountPaginationHeaders(t *testing.T) {
	for _, tc := range []struct {
		name   string
		header string
		value  string
	}{
		{name: "invalid page", header: "x-page", value: "abc"},
		{name: "negative page", header: "x-page", value: "-1"},
		{name: "invalid total pages", header: "x-total-pages", value: "abc"},
		{name: "negative total pages", header: "x-total-pages", value: "-1"},
		{name: "invalid total", header: "x-total", value: "abc"},
		{name: "negative total", header: "x-total", value: "-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			headers := http.Header{}
			headers.Set(tc.header, tc.value)

			data, pagination, errDetail := DecodeSearchList(json.RawMessage(`[{"id":"proj-1"}]`), headers, 20)

			require.Nil(t, data)
			require.Nil(t, pagination)
			require.NotNil(t, errDetail)
			require.Equal(t, "PAGINATION_INVALID", errDetail.Code)
			require.Equal(t, "general", errDetail.Category)
			require.Contains(t, errDetail.Message, tc.header)
			require.Contains(t, errDetail.Message, tc.value)
		})
	}
}

func TestDecodeSearchListWrapperUsesBodyPagination(t *testing.T) {
	headers := http.Header{}
	headers.Set("x-next-page", "header-token")
	headers.Set("x-per-page", "5")

	data, pagination, errDetail := DecodeSearchList(json.RawMessage(`{"data":[{"id":"tc-1"}],"nextPage":"body-token"}`), headers, 20)

	require.Nil(t, errDetail)
	require.Equal(t, []map[string]any{{"id": "tc-1"}}, data)
	require.NotNil(t, pagination.NextToken)
	require.Equal(t, "body-token", *pagination.NextToken)
	require.Equal(t, 5, pagination.PageSize)
	require.True(t, pagination.HasMore)
}

func TestDecodeSearchListRejectsInvalidShape(t *testing.T) {
	for _, body := range []string{`{"data":{}}`, `{"unexpected":true}`, `{"data":null}`} {
		t.Run(body, func(t *testing.T) {
			data, pagination, errDetail := DecodeSearchList(json.RawMessage(body), http.Header{}, 20)

			require.Nil(t, data)
			require.Nil(t, pagination)
			require.NotNil(t, errDetail)
			require.Equal(t, "RESPONSE_DECODE_FAILED", errDetail.Code)
			require.Equal(t, "general", errDetail.Category)
			require.False(t, errDetail.Retryable)
		})
	}
}

func TestDecodeSearchListRejectsBusinessErrorEnvelope(t *testing.T) {
	for _, tc := range []struct {
		name    string
		body    string
		message string
	}{
		{name: "success false", body: `{"success":false,"errorMessage":"search failed"}`, message: "search failed"},
		{name: "code message", body: `{"code":"InvalidParameter","message":"bad query"}`, message: "bad query"},
		{name: "capitalized", body: `{"Code":"InvalidParameter","Message":"bad query"}`, message: "bad query"},
		{name: "invalid success type", body: `{"success":"false","message":"bad success"}`, message: "bad success"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, pagination, errDetail := DecodeSearchList(json.RawMessage(tc.body), http.Header{}, 20)

			require.Nil(t, data)
			require.Nil(t, pagination)
			require.NotNil(t, errDetail)
			require.Equal(t, "UPSTREAM_BUSINESS_ERROR", errDetail.Code)
			require.Equal(t, "upstream", errDetail.Category)
			require.False(t, errDetail.Retryable)
			require.Contains(t, errDetail.Message, tc.message)
		})
	}
}
