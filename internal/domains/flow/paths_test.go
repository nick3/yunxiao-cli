package flow

import "testing"

func TestFlowReferencePathsUseRegionEndpoints(t *testing.T) {
	baseURL := "https://region.example.com"
	orgID := "org 123"

	tests := map[string]string{
		"pipelines": pipelinesPath(baseURL, orgID),
		"pipeline":  pipelinesPath(baseURL, orgID) + "/p%2F1",
		"runs":      pipelinesPath(baseURL, orgID) + "/p%2F1/runs",
		"run":       pipelinesPath(baseURL, orgID) + "/p%2F1/runs/r%2F1",
	}

	expected := map[string]string{
		"pipelines": "/oapi/v1/flow/pipelines",
		"pipeline":  "/oapi/v1/flow/pipelines/p%2F1",
		"runs":      "/oapi/v1/flow/pipelines/p%2F1/runs",
		"run":       "/oapi/v1/flow/pipelines/p%2F1/runs/r%2F1",
	}

	for name, got := range tests {
		if got != expected[name] {
			t.Fatalf("%s path = %q, want %q", name, got, expected[name])
		}
	}
}

func TestFlowReferencePathsUseCenterOrganizationEndpoints(t *testing.T) {
	baseURL := "https://openapi-rdc.aliyuncs.com"
	orgID := "org 123"

	tests := map[string]string{
		"pipelines": pipelinesPath(baseURL, orgID),
		"pipeline":  pipelinesPath(baseURL, orgID) + "/p%2F1",
		"runs":      pipelinesPath(baseURL, orgID) + "/p%2F1/runs",
		"run":       pipelinesPath(baseURL, orgID) + "/p%2F1/runs/r%2F1",
	}

	expected := map[string]string{
		"pipelines": "/oapi/v1/flow/organizations/org%20123/pipelines",
		"pipeline":  "/oapi/v1/flow/organizations/org%20123/pipelines/p%2F1",
		"runs":      "/oapi/v1/flow/organizations/org%20123/pipelines/p%2F1/runs",
		"run":       "/oapi/v1/flow/organizations/org%20123/pipelines/p%2F1/runs/r%2F1",
	}

	for name, got := range tests {
		if got != expected[name] {
			t.Fatalf("%s path = %q, want %q", name, got, expected[name])
		}
	}
}
