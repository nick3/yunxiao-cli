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

func TestDecodeSearchListRawArrayInvalidPerPageFallsBack(t *testing.T) {
	headers := http.Header{}
	headers.Set("x-next-page", "2")
	headers.Set("x-per-page", "invalid")

	_, pagination, errDetail := DecodeSearchList(json.RawMessage(`[{"id":"proj-1"}]`), headers, 20)

	require.Nil(t, errDetail)
	require.Equal(t, 20, pagination.PageSize)
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
	data, pagination, errDetail := DecodeSearchList(json.RawMessage(`{"data":{}}`), http.Header{}, 20)

	require.Nil(t, data)
	require.Nil(t, pagination)
	require.NotNil(t, errDetail)
	require.Equal(t, "RESPONSE_DECODE_FAILED", errDetail.Code)
	require.Equal(t, "general", errDetail.Category)
	require.False(t, errDetail.Retryable)
}
