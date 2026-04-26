package shared

import "strconv"

// SearchResponse is the common envelope returned by Yunxiao `:search` style endpoints.
type SearchResponse struct {
	Data     []map[string]any `json:"data"`
	NextPage any              `json:"nextPage"`
}

// ApplyPageToken sets the "page" key on a search payload, preferring an int when the token parses cleanly.
func ApplyPageToken(payload map[string]any, pageToken string) {
	if pageToken == "" {
		return
	}
	if page, err := strconv.Atoi(pageToken); err == nil {
		payload["page"] = page
		return
	}
	payload["page"] = pageToken
}
