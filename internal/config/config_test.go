package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestGetOrganizationIDUsesRegionDefaultOrgBeforeConfigInRegionMode(t *testing.T) {
	t.Setenv("YUNXIAO_API_BASE_URL", "https://devops.aliyuncs.com")
	t.Setenv("YUNXIAO_REGION_DEFAULT_ORG_ID", "org-from-region")
	viper.Set("organization_id", "org-from-config")
	t.Cleanup(func() { viper.Reset() })

	orgID := GetOrganizationID("")

	require.Equal(t, "org-from-region", orgID)
}

func TestGetOrganizationIDUsesConfigOrgInCenterMode(t *testing.T) {
	t.Setenv("YUNXIAO_API_BASE_URL", "https://openapi-rdc.aliyuncs.com")
	t.Setenv("YUNXIAO_REGION_DEFAULT_ORG_ID", "org-from-region")
	viper.Set("organization_id", "org-from-config")
	t.Cleanup(func() { viper.Reset() })

	orgID := GetOrganizationID("")

	require.Equal(t, "org-from-config", orgID)
}
