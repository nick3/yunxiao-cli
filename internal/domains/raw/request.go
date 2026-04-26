package raw

import (
	"context"
	"net/http"
	"strings"

	"github.com/aliyun/yunxiao-cli/internal/domains/shared"
	"github.com/aliyun/yunxiao-cli/internal/httpx"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
)

func Validate(method, path string) *output.ErrorDetail {
	if strings.ToUpper(method) != http.MethodGet {
		return &output.ErrorDetail{Code: "PARAM_INVALID", Category: "param", Retryable: false, Message: "raw request only supports GET in this phase"}
	}
	if !strings.HasPrefix(path, "/oapi/") {
		return &output.ErrorDetail{Code: "PARAM_INVALID", Category: "param", Retryable: false, Message: "path must start with /oapi/"}
	}
	return nil
}

func Request(ctx context.Context, client *httpx.Client, method, path string) (any, *output.ErrorDetail) {
	if errDetail := Validate(method, path); errDetail != nil {
		return nil, errDetail
	}
	var data any
	if errDetail := shared.RequestJSON(ctx, client, http.MethodGet, path, &data); errDetail != nil {
		return nil, errDetail
	}
	return data, nil
}
