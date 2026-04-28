package projex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/nick3/yunxiao-cli/internal/domains/shared"
	"github.com/nick3/yunxiao-cli/internal/httpx"
	"github.com/nick3/yunxiao-cli/internal/model/output"
)

func ListAllWorkitemTypes(ctx context.Context, client *httpx.Client, organizationID string) ([]map[string]any, *output.ErrorDetail) {
	return requestWorkitemTypeList(ctx, client, workitemTypesPath(client.BaseURL, organizationID))
}

func ListProjectWorkitemTypes(ctx context.Context, client *httpx.Client, organizationID, projectID, category string) ([]map[string]any, *output.ErrorDetail) {
	path := projectWorkitemTypesPath(client.BaseURL, organizationID, projectID)
	if category != "" {
		query := url.Values{"category": []string{category}}
		path += "?" + query.Encode()
	}
	return requestWorkitemTypeList(ctx, client, path)
}

func GetWorkitemType(ctx context.Context, client *httpx.Client, organizationID, typeID string) (map[string]any, *output.ErrorDetail) {
	var body json.RawMessage
	path := workitemTypesPath(client.BaseURL, organizationID) + "/" + url.PathEscape(typeID)
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path, &body); errDetail != nil {
		return nil, errDetail
	}
	return decodeObjectOrResult(body)
}

func ListWorkitemRelationTypes(ctx context.Context, client *httpx.Client, organizationID, typeID, relationType string) ([]map[string]any, *output.ErrorDetail) {
	path := workitemTypesPath(client.BaseURL, organizationID) + "/" + url.PathEscape(typeID) + "/relationWorkitemTypes"
	if relationType != "" {
		query := url.Values{"relationType": []string{relationType}}
		path += "?" + query.Encode()
	}
	return requestWorkitemTypeList(ctx, client, path)
}

func GetWorkitemTypeFields(ctx context.Context, client *httpx.Client, organizationID, projectID, typeID string) (map[string]any, *output.ErrorDetail) {
	var body json.RawMessage
	path := projectWorkitemTypesPath(client.BaseURL, organizationID, projectID) + "/" + url.PathEscape(typeID) + "/fields"
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path, &body); errDetail != nil {
		return nil, errDetail
	}
	return decodeObjectOrResult(body)
}

func GetWorkitemTypeWorkflow(ctx context.Context, client *httpx.Client, organizationID, projectID, typeID string) (map[string]any, *output.ErrorDetail) {
	var body json.RawMessage
	path := projectWorkitemTypesPath(client.BaseURL, organizationID, projectID) + "/" + url.PathEscape(typeID) + "/workflows"
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path, &body); errDetail != nil {
		return nil, errDetail
	}
	return decodeObjectOrResult(body)
}

func requestWorkitemTypeList(ctx context.Context, client *httpx.Client, path string) ([]map[string]any, *output.ErrorDetail) {
	var body json.RawMessage
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path, &body); errDetail != nil {
		return nil, errDetail
	}
	return decodeArrayOrResult(body, "workitem metadata")
}

func decodeArrayOrResult(body json.RawMessage, resourceName string) ([]map[string]any, *output.ErrorDetail) {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return nil, &output.ErrorDetail{Code: "EMPTY_RESPONSE", Category: "general", Retryable: false, Message: "upstream returned an empty response body"}
	}
	if body[0] == '[' {
		var data []map[string]any
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, decodeWorkitemResponseError(err, resourceName)
		}
		return data, nil
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, decodeWorkitemResponseError(err, resourceName)
	}
	var envelope map[string]any
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, decodeWorkitemResponseError(err, resourceName)
	}
	if errDetail := detectBusinessError(envelope); errDetail != nil {
		return nil, errDetail
	}
	result, ok := raw["result"]
	if !ok {
		return nil, unexpectedArrayOrResultResponse(resourceName)
	}
	var data []map[string]any
	if err := json.Unmarshal(result, &data); err != nil || data == nil {
		return nil, unexpectedArrayOrResultResponse(resourceName)
	}
	return data, nil
}

func unexpectedArrayOrResultResponse(resourceName string) *output.ErrorDetail {
	return &output.ErrorDetail{Code: "RESPONSE_DECODE_FAILED", Category: "general", Retryable: false, Message: "failed to decode " + resourceName + " response: expected array or object with result array"}
}

func decodeObjectOrResult(body json.RawMessage, resourceNames ...string) (map[string]any, *output.ErrorDetail) {
	resourceName := "workitem metadata"
	if len(resourceNames) > 0 && strings.TrimSpace(resourceNames[0]) != "" {
		resourceName = resourceNames[0]
	}
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return nil, &output.ErrorDetail{Code: "EMPTY_RESPONSE", Category: "general", Retryable: false, Message: "upstream returned an empty response body"}
	}
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, decodeWorkitemResponseError(err, resourceName)
	}
	if data == nil {
		return nil, &output.ErrorDetail{Code: "RESPONSE_DECODE_FAILED", Category: "general", Retryable: false, Message: "failed to decode " + resourceName + " response: expected object"}
	}
	if errDetail := detectBusinessError(data); errDetail != nil {
		return nil, errDetail
	}
	if result, ok := data["result"]; ok {
		object, ok := result.(map[string]any)
		if !ok || object == nil {
			return nil, &output.ErrorDetail{Code: "RESPONSE_DECODE_FAILED", Category: "general", Retryable: false, Message: "failed to decode " + resourceName + " response: result must be an object"}
		}
		if errDetail := detectBusinessError(object); errDetail != nil {
			return nil, errDetail
		}
		return object, nil
	}
	return data, nil
}

func decodeResourceObjectOrResult(body json.RawMessage, resourceName string, identityKeys ...string) (map[string]any, *output.ErrorDetail) {
	data, errDetail := decodeObjectOrResult(body, resourceName)
	if errDetail != nil {
		return nil, errDetail
	}
	if !hasResourceIdentity(data, identityKeys...) {
		return nil, &output.ErrorDetail{Code: "RESPONSE_DECODE_FAILED", Category: "general", Retryable: false, Message: "failed to decode " + resourceName + " response: expected resource object with id"}
	}
	return data, nil
}

func decodeUpdateConfirmationOrResult(body json.RawMessage, resourceName string, identityKeys ...string) *output.ErrorDetail {
	data, errDetail := decodeObjectOrResult(body, resourceName)
	if errDetail != nil {
		return errDetail
	}
	if hasResourceIdentity(data, identityKeys...) || isExplicitSuccessOnly(data) {
		return nil
	}
	return &output.ErrorDetail{Code: "RESPONSE_DECODE_FAILED", Category: "general", Retryable: false, Message: "failed to decode " + resourceName + " response: expected resource object with id or explicit success confirmation"}
}

func isExplicitSuccessOnly(data map[string]any) bool {
	success, ok := data["success"].(bool)
	if !ok || !success {
		return false
	}
	for key := range data {
		if key != "success" && !isSuccessMetadataKey(key) {
			return false
		}
	}
	return true
}

func isSuccessMetadataKey(key string) bool {
	switch key {
	case "requestId", "RequestId", "requestID", "RequestID", "traceId", "TraceId", "traceID", "TraceID":
		return true
	default:
		return false
	}
}

func hasResourceIdentity(data map[string]any, keys ...string) bool {
	for _, key := range keys {
		if value, ok := data[key].(string); ok && strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

func detectBusinessError(data map[string]any) *output.ErrorDetail {
	if value, ok := data["success"]; ok {
		success, ok := value.(bool)
		if !ok || !success {
			return upstreamBusinessError(data)
		}
	}
	if hasStringValue(data, "errorCode") || hasStringValue(data, "errorMessage") || hasStringValue(data, "Code") || hasStringValue(data, "Message") {
		return upstreamBusinessError(data)
	}
	if hasStringValue(data, "code") && (hasStringValue(data, "message") || hasStringValue(data, "msg")) {
		return upstreamBusinessError(data)
	}
	return nil
}

func hasStringValue(data map[string]any, key string) bool {
	value, ok := data[key].(string)
	return ok && strings.TrimSpace(value) != ""
}

func upstreamBusinessError(data map[string]any) *output.ErrorDetail {
	message := "upstream returned a business error"
	for _, key := range []string{"errorMessage", "message", "msg", "Message"} {
		if value, ok := data[key].(string); ok && strings.TrimSpace(value) != "" {
			message = value
			break
		}
	}
	return &output.ErrorDetail{Code: "UPSTREAM_BUSINESS_ERROR", Category: "upstream", Retryable: false, Message: message}
}

func decodeWorkitemResponseError(err error, resourceName string) *output.ErrorDetail {
	return &output.ErrorDetail{Code: "RESPONSE_DECODE_FAILED", Category: "general", Retryable: false, Message: fmt.Sprintf("failed to decode "+resourceName+" response: %v", err)}
}

func workitemTypesPath(baseURL, organizationID string) string {
	if shared.IsRegionBaseURL(baseURL) {
		return "/oapi/v1/projex/workitemTypes"
	}
	return "/oapi/v1/projex/organizations/" + url.PathEscape(organizationID) + "/workitemTypes"
}

func projectWorkitemTypesPath(baseURL, organizationID, projectID string) string {
	if shared.IsRegionBaseURL(baseURL) {
		return "/oapi/v1/projex/projects/" + url.PathEscape(projectID) + "/workitemTypes"
	}
	return "/oapi/v1/projex/organizations/" + url.PathEscape(organizationID) + "/projects/" + url.PathEscape(projectID) + "/workitemTypes"
}
