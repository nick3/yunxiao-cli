package org

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aliyun/yunxiao-cli/internal/domains/shared"
	"github.com/aliyun/yunxiao-cli/internal/httpx"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
)

type CurrentUser struct {
	Data               map[string]any
	UserID             string
	UserName           string
	LastOrganizationID string
}

func GetCurrentUser(ctx context.Context, client *httpx.Client) (*CurrentUser, *output.ErrorDetail) {
	data := map[string]any{}
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, "/oapi/v1/platform/user", &data); errDetail != nil {
		return nil, errDetail
	}

	userID := stringField(data, "id")
	userName := stringField(data, "name")
	lastOrganizationID := stringField(data, "lastOrganization")
	if userID != "" {
		data["userId"] = userID
	}
	if userName != "" {
		data["userName"] = userName
	}
	if lastOrganizationID != "" {
		data["lastOrganizationId"] = lastOrganizationID
	}

	return &CurrentUser{Data: data, UserID: userID, UserName: userName, LastOrganizationID: lastOrganizationID}, nil
}

func stringField(data map[string]any, key string) string {
	value, ok := data[key]
	if !ok || value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return s
	}
	return fmt.Sprint(value)
}
