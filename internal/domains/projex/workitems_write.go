package projex

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"net/url"

	"github.com/nick3/yunxiao-cli/internal/domains/shared"
	"github.com/nick3/yunxiao-cli/internal/httpx"
	"github.com/nick3/yunxiao-cli/internal/model/output"
)

type WorkitemCreateInput struct {
	AssignedTo        string
	SpaceID           string
	Subject           string
	WorkitemTypeID    string
	CustomFieldValues map[string]any
	Description       string
	FormatType        string
	Labels            []string
	ParentID          string
	Participants      []string
	Sprint            string
	Trackers          []string
	Verifier          string
	Versions          []string
}

type WorkitemUpdateInput struct {
	Subject           string
	Description       string
	FormatType        string
	Status            string
	AssignedTo        string
	Priority          string
	Labels            []string
	Sprint            string
	Trackers          []string
	Verifier          string
	Participants      []string
	Versions          []string
	CustomFieldValues map[string]any
}

type WorkitemUpdateResult struct {
	WorkitemID string `json:"workitem_id"`
	Updated    bool   `json:"updated"`
}

func CreateWorkitem(ctx context.Context, client *httpx.Client, organizationID string, input WorkitemCreateInput) (map[string]any, *output.ErrorDetail) {
	payload := map[string]any{
		"assignedTo":     input.AssignedTo,
		"spaceId":        input.SpaceID,
		"subject":        input.Subject,
		"workitemTypeId": input.WorkitemTypeID,
	}
	if len(input.CustomFieldValues) > 0 {
		payload["customFieldValues"] = input.CustomFieldValues
	}
	if input.Description != "" {
		payload["description"] = input.Description
	}
	if input.FormatType != "" {
		payload["formatType"] = input.FormatType
	}
	if len(input.Labels) > 0 {
		payload["labels"] = input.Labels
	}
	if input.ParentID != "" {
		payload["parentId"] = input.ParentID
	}
	if len(input.Participants) > 0 {
		payload["participants"] = input.Participants
	}
	if input.Sprint != "" {
		payload["sprint"] = input.Sprint
	}
	if len(input.Trackers) > 0 {
		payload["trackers"] = input.Trackers
	}
	if input.Verifier != "" {
		payload["verifier"] = input.Verifier
	}
	if len(input.Versions) > 0 {
		payload["versions"] = input.Versions
	}

	var body json.RawMessage
	if errDetail := shared.RequestJSONWithBody(ctx, client, http.MethodPost, workitemsPath(client.BaseURL, organizationID), payload, &body); errDetail != nil {
		return nil, errDetail
	}
	return decodeResourceObjectOrResult(body, "workitem create", "id", "identifier", "workitemId", "workItemId")
}

func UpdateWorkitem(ctx context.Context, client *httpx.Client, organizationID, workitemID string, input WorkitemUpdateInput) (*WorkitemUpdateResult, *output.ErrorDetail) {
	payload := map[string]any{}
	if input.Subject != "" {
		payload["subject"] = input.Subject
	}
	if input.Description != "" {
		payload["description"] = input.Description
	}
	if input.FormatType != "" {
		payload["formatType"] = input.FormatType
	}
	if input.Status != "" {
		payload["status"] = input.Status
	}
	if input.AssignedTo != "" {
		payload["assignedTo"] = input.AssignedTo
	}
	if input.Priority != "" {
		payload["priority"] = input.Priority
	}
	if len(input.Labels) > 0 {
		payload["labels"] = input.Labels
	}
	if input.Sprint != "" {
		payload["sprint"] = input.Sprint
	}
	if len(input.Trackers) > 0 {
		payload["trackers"] = input.Trackers
	}
	if input.Verifier != "" {
		payload["verifier"] = input.Verifier
	}
	if len(input.Participants) > 0 {
		payload["participants"] = input.Participants
	}
	if len(input.Versions) > 0 {
		payload["versions"] = input.Versions
	}
	maps.Copy(payload, input.CustomFieldValues)
	if len(payload) == 0 {
		return nil, &output.ErrorDetail{Code: "PARAM_REQUIRED", Category: "param", Retryable: false, Message: "at least one update field is required"}
	}

	path := workitemsPath(client.BaseURL, organizationID) + "/" + url.PathEscape(workitemID)
	var body json.RawMessage
	if errDetail := shared.RequestJSONWithBody(ctx, client, http.MethodPut, path, payload, &body); errDetail != nil {
		return nil, errDetail
	}
	if errDetail := decodeUpdateConfirmationOrResult(body, "workitem update", "id", "identifier", "workitemId", "workItemId"); errDetail != nil {
		return nil, errDetail
	}
	return &WorkitemUpdateResult{WorkitemID: workitemID, Updated: true}, nil
}
