package validation

import "github.com/nick3/yunxiao-cli/internal/model/output"

func PageSize(pageSize int) *output.ErrorDetail {
	if pageSize > 0 {
		return nil
	}
	return &output.ErrorDetail{Code: "PARAM_INVALID", Category: "param", Retryable: false, Message: "page_size must be greater than 0"}
}
