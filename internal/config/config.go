package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

func Init() error {
	viper.SetEnvPrefix("YUNXIAO")
	viper.AutomaticEnv()

	viper.SetDefault("api_base_url", "https://devops.aliyuncs.com")
	viper.SetDefault("region", "")
	viper.SetDefault("timeout", 30)

	configPath := os.Getenv("YUNXIAO_CONFIG_FILE")
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to resolve user home directory: %w", err)
		}
		configPath = home + "/.yunxiao"
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return nil
		}
		return fmt.Errorf("failed to read config: %w", err)
	}
	return nil
}

func GetBaseURL() string {
	if url := os.Getenv("YUNXIAO_API_BASE_URL"); url != "" {
		return url
	}
	return viper.GetString("api_base_url")
}

func GetTimeout(flagVal int, flagChanged bool) int {
	if flagChanged {
		return flagVal
	}
	if envVal := os.Getenv("YUNXIAO_TIMEOUT"); envVal != "" {
		if timeout, err := strconv.Atoi(envVal); err == nil {
			return timeout
		}
	}
	return viper.GetInt("timeout")
}

func GetOrganizationID(flagVal string) string {
	if flagVal != "" {
		return flagVal
	}
	if envVal := os.Getenv("YUNXIAO_ORGANIZATION_ID"); envVal != "" {
		return envVal
	}
	if defaultOrgID := os.Getenv("YUNXIAO_REGION_DEFAULT_ORG_ID"); defaultOrgID != "" && IsRegionBaseURL(GetBaseURL()) {
		return defaultOrgID
	}
	return viper.GetString("organization_id")
}

func IsRegionBaseURL(baseURL string) bool {
	return !strings.Contains(baseURL, "openapi-rdc.aliyuncs.com")
}
