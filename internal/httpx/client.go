package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aliyun/yunxiao-cli/internal/model/output"
)

type Client struct {
	HTTP           *http.Client
	BaseURL        string
	Token          string
	TraceID        string
	MaxRetries     int
	NoRetry        bool
	RequestTimeout time.Duration
	Quiet          bool
}

func NewClient(baseURL, token string, timeoutSec int, noRetry bool, traceID string) *Client {
	return &Client{
		HTTP:           &http.Client{Timeout: time.Duration(timeoutSec) * time.Second},
		BaseURL:        baseURL,
		Token:          token,
		TraceID:        traceID,
		MaxRetries:     3,
		NoRetry:        noRetry,
		RequestTimeout: time.Duration(timeoutSec) * time.Second,
	}
}

func (c *Client) Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	var lastErr error
	attempts := 1
	if !c.NoRetry && isRetryableMethod(method) {
		attempts = c.MaxRetries + 1
	}
	for i := range attempts {
		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("x-yunxiao-token", c.Token)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		if c.TraceID != "" {
			req.Header.Set("X-Trace-Id", c.TraceID)
		}
		resp, err := c.HTTP.Do(req)
		if err != nil {
			if os.IsTimeout(err) {
				lastErr = fmt.Errorf("request timed out after %s", c.RequestTimeout)
			} else {
				lastErr = err
			}
			continue
		}
		if !c.NoRetry && isRetryable(resp.StatusCode) && i < attempts-1 {
			resp.Body.Close()
			backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
			if ra := parseRetryAfter(resp.Header.Get("Retry-After")); ra > 0 {
				backoff = ra
			}
			if !c.Quiet {
				fmt.Fprintf(os.Stderr, "[RETRY] upstream returned %d, attempt %d/%d\n", resp.StatusCode, i+1, c.MaxRetries)
			}
			time.Sleep(backoff)
			lastErr = fmt.Errorf("upstream returned %d", resp.StatusCode)
			continue
		}
		return resp, nil
	}
	return nil, lastErr
}

func isRetryable(status int) bool {
	return status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
}

func isRetryableMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

func parseRetryAfter(val string) time.Duration {
	if val == "" {
		return 0
	}
	secs, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return time.Duration(secs) * time.Second
}

func ClassifyHTTPError(statusCode int, respBody []byte) *output.ErrorDetail {
	s := statusCode
	switch {
	case statusCode == http.StatusUnauthorized:
		return &output.ErrorDetail{Code: "AUTH_FAILED", Category: "auth", Retryable: false, Message: "authentication failed", UpstreamStatus: &s}
	case statusCode == http.StatusForbidden:
		return &output.ErrorDetail{Code: "FORBIDDEN", Category: "forbidden", Retryable: false, Message: "permission denied", UpstreamStatus: &s}
	case statusCode == http.StatusNotFound:
		return &output.ErrorDetail{Code: "RESOURCE_NOT_FOUND", Category: "not_found", Retryable: false, Message: "resource not found", UpstreamStatus: &s}
	case statusCode == http.StatusTooManyRequests:
		return &output.ErrorDetail{Code: "RATE_LIMITED", Category: "rate_limit", Retryable: true, Message: "API rate limit exceeded", UpstreamStatus: &s}
	case statusCode >= 500:
		return &output.ErrorDetail{Code: "UPSTREAM_UNAVAILABLE", Category: "upstream", Retryable: true, Message: "upstream service error", UpstreamStatus: &s}
	default:
		msg := "request failed"
		if len(respBody) > 0 {
			var parsed map[string]interface{}
			if json.Unmarshal(respBody, &parsed) == nil {
				if m, ok := parsed["message"].(string); ok {
					msg = m
				}
			}
		}
		return &output.ErrorDetail{Code: "REQUEST_FAILED", Category: "general", Retryable: false, Message: msg, UpstreamStatus: &s}
	}
}
