package projex

import "testing"

func TestProjexReferencePathsUseRegionEndpoints(t *testing.T) {
	baseURL := "https://region.example.com"
	orgID := "org-123"

	tests := map[string]string{
		"projects":         projectsPath(baseURL, orgID),
		"project":          projectsPath(baseURL, orgID) + "/proj-1",
		"workitems_search": workitemsPath(baseURL, orgID) + ":search",
		"workitem":         workitemsPath(baseURL, orgID) + "/wi-1",
		"project_sprints":  sprintsPath(baseURL, orgID, "proj-1"),
	}

	expected := map[string]string{
		"projects":         "/oapi/v1/projex/projects",
		"project":          "/oapi/v1/projex/projects/proj-1",
		"workitems_search": "/oapi/v1/projex/workitems:search",
		"workitem":         "/oapi/v1/projex/workitems/wi-1",
		"project_sprints":  "/oapi/v1/projex/projects/proj-1/sprints",
	}

	for name, got := range tests {
		if got != expected[name] {
			t.Fatalf("%s path = %q, want %q", name, got, expected[name])
		}
	}
}

func TestProjexReferencePathsUseCenterOrganizationEndpoints(t *testing.T) {
	baseURL := "https://openapi-rdc.aliyuncs.com"
	orgID := "org-123"

	tests := map[string]string{
		"projects":         projectsPath(baseURL, orgID),
		"project":          projectsPath(baseURL, orgID) + "/proj-1",
		"workitems_search": workitemsPath(baseURL, orgID) + ":search",
		"workitem":         workitemsPath(baseURL, orgID) + "/wi-1",
		"project_sprints":  sprintsPath(baseURL, orgID, "proj-1"),
	}

	expected := map[string]string{
		"projects":         "/oapi/v1/projex/organizations/org-123/projects",
		"project":          "/oapi/v1/projex/organizations/org-123/projects/proj-1",
		"workitems_search": "/oapi/v1/projex/organizations/org-123/workitems:search",
		"workitem":         "/oapi/v1/projex/organizations/org-123/workitems/wi-1",
		"project_sprints":  "/oapi/v1/projex/organizations/org-123/projects/proj-1/sprints",
	}

	for name, got := range tests {
		if got != expected[name] {
			t.Fatalf("%s path = %q, want %q", name, got, expected[name])
		}
	}
}
