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

func TestDecodeSearchListWrapperUsesBodyPagination(t *testing.T) {
	headers := http.Header{}
	headers.Set("x-next-page", "header-token")
	headers.Set("x-per-page", "5")

	data, pagination, errDetail := DecodeSearchList(json.RawMessage(`{"data":[{"id":"tc-1"}],"nextPage":"body-token"}`), headers, 20)

	require.Nil(t, errDetail)
	require.Equal(t, []map[string]any{{"id": "tc-1"}}, data)
	require.NotNil(t, pagination.NextToken)
	require.Equal(t, "body-token", *pagination.NextToken)
	require.Equal(t, 20, pagination.PageSize)
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
