package org

import "testing"

func TestOrgReferencePathsUseRegionEndpoints(t *testing.T) {
	got := membersPath("https://region.example.com", "org 123")
	want := "/oapi/v1/platform/members"
	if got != want {
		t.Fatalf("members path = %q, want %q", got, want)
	}
}

func TestOrgReferencePathsUseCenterOrganizationEndpoints(t *testing.T) {
	got := membersPath("https://openapi-rdc.aliyuncs.com", "org 123")
	want := "/oapi/v1/platform/organizations/org%20123/members"
	if got != want {
		t.Fatalf("members path = %q, want %q", got, want)
	}
}
