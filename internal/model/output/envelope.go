package output

type Envelope struct {
	Version string       `json:"version"`
	Data    interface{}  `json:"data"`
	Meta    *Meta        `json:"meta"`
	Error   *ErrorDetail `json:"error"`
}

type Meta struct {
	TraceID    string      `json:"trace_id,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	NextToken *string `json:"next_token"`
	PageSize  int     `json:"page_size"`
	HasMore   bool    `json:"has_more"`
}

type ErrorDetail struct {
	Code           string `json:"code"`
	Category       string `json:"category"`
	Retryable      bool   `json:"retryable"`
	Message        string `json:"message"`
	UpstreamStatus *int   `json:"upstream_status"`
}
