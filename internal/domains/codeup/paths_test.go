package codeup

import "testing"

func TestCodeupReferencePathsUseRegionEndpoints(t *testing.T) {
	baseURL := "https://region.example.com"
	orgID := "org 123"

	tests := map[string]string{
		"repositories": repositoriesPath(baseURL, orgID),
		"repository":   repositoryPath(baseURL, orgID, "repo/1"),
		"branches":     repositoryPath(baseURL, orgID, "repo/1") + "/branches",
		"commits":      repositoryPath(baseURL, orgID, "repo/1") + "/commits",
		"files":        repositoryPath(baseURL, orgID, "repo/1") + "/files/README.md",
		"compares":     repositoryPath(baseURL, orgID, "repo/1") + "/compares",
	}

	expected := map[string]string{
		"repositories": "/oapi/v1/codeup/repositories",
		"repository":   "/oapi/v1/codeup/repositories/repo%2F1",
		"branches":     "/oapi/v1/codeup/repositories/repo%2F1/branches",
		"commits":      "/oapi/v1/codeup/repositories/repo%2F1/commits",
		"files":        "/oapi/v1/codeup/repositories/repo%2F1/files/README.md",
		"compares":     "/oapi/v1/codeup/repositories/repo%2F1/compares",
	}

	for name, got := range tests {
		if got != expected[name] {
			t.Fatalf("%s path = %q, want %q", name, got, expected[name])
		}
	}
}

func TestCodeupReferencePathsUseCenterOrganizationEndpoints(t *testing.T) {
	baseURL := "https://openapi-rdc.aliyuncs.com"
	orgID := "org 123"

	tests := map[string]string{
		"repositories": repositoriesPath(baseURL, orgID),
		"repository":   repositoryPath(baseURL, orgID, "repo/1"),
		"branches":     repositoryPath(baseURL, orgID, "repo/1") + "/branches",
		"commits":      repositoryPath(baseURL, orgID, "repo/1") + "/commits",
		"files":        repositoryPath(baseURL, orgID, "repo/1") + "/files/README.md",
		"compares":     repositoryPath(baseURL, orgID, "repo/1") + "/compares",
	}

	expected := map[string]string{
		"repositories": "/oapi/v1/codeup/organizations/org%20123/repositories",
		"repository":   "/oapi/v1/codeup/organizations/org%20123/repositories/repo%2F1",
		"branches":     "/oapi/v1/codeup/organizations/org%20123/repositories/repo%2F1/branches",
		"commits":      "/oapi/v1/codeup/organizations/org%20123/repositories/repo%2F1/commits",
		"files":        "/oapi/v1/codeup/organizations/org%20123/repositories/repo%2F1/files/README.md",
		"compares":     "/oapi/v1/codeup/organizations/org%20123/repositories/repo%2F1/compares",
	}

	for name, got := range tests {
		if got != expected[name] {
			t.Fatalf("%s path = %q, want %q", name, got, expected[name])
		}
	}
}
