package testhub

import "testing"

func TestTesthubReferencePathsUseRegionEndpoints(t *testing.T) {
	baseURL := "https://region.example.com"
	orgID := "org-123"

	tests := map[string]string{
		"testcases_search": testcasesPath(baseURL, orgID, "repo-1") + ":search",
		"testcase":         testcasesPath(baseURL, orgID, "repo-1") + "/tc-1",
		"directories":      directoriesPath(baseURL, orgID, "repo-1"),
		"testplans":        testplansPath(baseURL, orgID),
	}

	expected := map[string]string{
		"testcases_search": "/oapi/v1/testhub/testRepos/repo-1/testcases:search",
		"testcase":         "/oapi/v1/testhub/testRepos/repo-1/testcases/tc-1",
		"directories":      "/oapi/v1/testhub/testRepos/repo-1/directories",
		"testplans":        "/oapi/v1/projex/testPlan/list",
	}

	for name, got := range tests {
		if got != expected[name] {
			t.Fatalf("%s path = %q, want %q", name, got, expected[name])
		}
	}
}

func TestTesthubReferencePathsUseCenterOrganizationEndpoints(t *testing.T) {
	baseURL := "https://openapi-rdc.aliyuncs.com"
	orgID := "org-123"

	tests := map[string]string{
		"testcases_search": testcasesPath(baseURL, orgID, "repo-1") + ":search",
		"testcase":         testcasesPath(baseURL, orgID, "repo-1") + "/tc-1",
		"directories":      directoriesPath(baseURL, orgID, "repo-1"),
		"testplans":        testplansPath(baseURL, orgID),
	}

	expected := map[string]string{
		"testcases_search": "/oapi/v1/testhub/organizations/org-123/testRepos/repo-1/testcases:search",
		"testcase":         "/oapi/v1/testhub/organizations/org-123/testRepos/repo-1/testcases/tc-1",
		"directories":      "/oapi/v1/testhub/organizations/org-123/testRepos/repo-1/directories",
		"testplans":        "/oapi/v1/projex/organizations/org-123/testPlan/list",
	}

	for name, got := range tests {
		if got != expected[name] {
			t.Fatalf("%s path = %q, want %q", name, got, expected[name])
		}
	}
}
