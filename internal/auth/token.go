package auth

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func GetAccessToken() (string, error) {
	if token := os.Getenv("YUNXIAO_ACCESS_TOKEN"); token != "" {
		return token, nil
	}
	if token := viper.GetString("access_token"); token != "" {
		return token, nil
	}
	return "", fmt.Errorf("access token is not configured; set YUNXIAO_ACCESS_TOKEN or run `yunxiao auth`")
}
