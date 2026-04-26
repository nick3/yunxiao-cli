package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aliyun/yunxiao-cli/internal/httpx"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
)

func RequestJSON(ctx context.Context, client *httpx.Client, method, path string, target any) *output.ErrorDetail {
	_, errDetail := RequestJSONWithBodyAndHeaders(ctx, client, method, path, nil, target)
	return errDetail
}

func RequestJSONWithBody(ctx context.Context, client *httpx.Client, method, path string, payload any, target any) *output.ErrorDetail {
	_, errDetail := RequestJSONWithBodyAndHeaders(ctx, client, method, path, payload, target)
	return errDetail
}

func RequestJSONWithBodyAndHeaders(ctx context.Context, client *httpx.Client, method, path string, payload any, target any) (http.Header, *output.ErrorDetail) {
	var bodyReader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, &output.ErrorDetail{Code: "REQUEST_ENCODE_FAILED", Category: "general", Retryable: false, Message: err.Error()}
		}
		bodyReader = bytes.NewReader(body)
	}

	resp, err := client.Do(ctx, method, path, bodyReader)
	if err != nil {
		code := "NETWORK_ERROR"
		if containsTimeout(err.Error()) {
			code = "REQUEST_TIMEOUT"
		}
		return nil, &output.ErrorDetail{Code: code, Category: "network", Retryable: true, Message: err.Error()}
	}
	defer resp.Body.Close()

	headers := resp.Header.Clone()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return headers, &output.ErrorDetail{Code: "RESPONSE_READ_FAILED", Category: "general", Retryable: false, Message: readErr.Error()}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return headers, httpx.ClassifyHTTPError(resp.StatusCode, body)
	}
	if target == nil {
		return headers, nil
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return headers, &output.ErrorDetail{Code: "EMPTY_RESPONSE", Category: "general", Retryable: false, Message: "upstream returned an empty response body for a JSON request"}
	}
	if err := json.Unmarshal(body, target); err != nil {
		return headers, &output.ErrorDetail{Code: "RESPONSE_DECODE_FAILED", Category: "general", Retryable: false, Message: fmt.Sprintf("failed to decode response: %v", err)}
	}
	return headers, nil
}

func IsRegionBaseURL(baseURL string) bool {
	return !strings.Contains(baseURL, "openapi-rdc.aliyuncs.com")
}

func StringToken(value any) *string {
	switch v := value.(type) {
	case nil:
		return nil
	case string:
		if v == "" {
			return nil
		}
		return &v
	case float64:
		s := fmt.Sprintf("%.0f", v)
		return &s
	default:
		s := fmt.Sprint(v)
		if s == "" || s == "<nil>" {
			return nil
		}
		return &s
	}
}

func containsTimeout(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "timeout") || strings.Contains(lower, "timed out")
}
