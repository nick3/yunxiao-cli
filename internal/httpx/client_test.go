package httpx

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestDoRetriesHeadRequests(t *testing.T) {
	attempts := 0
	client := NewClient("https://example.test", "token", 30, false, "")
	client.Quiet = true
	client.HTTP = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		require.Equal(t, http.MethodHead, req.Method)
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    req,
		}, nil
	})}

	resp, err := client.Do(context.Background(), http.MethodHead, "/oapi/v1/test", nil)

	require.NoError(t, err)
	require.Equal(t, http.StatusBadGateway, resp.StatusCode)
	require.Equal(t, 4, attempts)
}

func TestDoDoesNotRetryPostRequests(t *testing.T) {
	attempts := 0
	client := NewClient("https://example.test", "token", 30, false, "")
	client.HTTP = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		require.Equal(t, http.MethodPost, req.Method)
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"message":"unavailable"}`)),
			Request:    req,
		}, nil
	})}

	resp, err := client.Do(context.Background(), http.MethodPost, "/oapi/v1/test", strings.NewReader(`{}`))

	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	require.Equal(t, 1, attempts)
}
