package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers *http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	token := strings.Replace(authHeader, "ApiKey ", "", 1)
	if token == "" {
		return "", fmt.Errorf("no api key provided")
	}
	return token, nil
}
