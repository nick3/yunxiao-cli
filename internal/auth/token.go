package auth

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func GetAccessToken() (string, error) {
	token := viper.GetString("access_token")
	if token == "" {
		token = os.Getenv("YUNXIAO_ACCESS_TOKEN")
	}
	if token == "" {
		return "", fmt.Errorf("YUNXIAO_ACCESS_TOKEN is missing or invalid")
	}
	return token, nil
}
