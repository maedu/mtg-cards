package auth

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/oauth2/v1"
)

var httpClient = &http.Client{}

func GetUserIDFromAccessToken(c *gin.Context, sendErrorStatus bool) (string, bool) {

	if token, ok := getAccessToken(c); ok {
		info, err := verifyAccessToken(token)
		if err != nil {

			if sendErrorStatus {
				c.JSON(http.StatusForbidden, "Token verification failed")
			}
			return "", false
		}

		return info.Email, true

	}
	if sendErrorStatus {
		c.JSON(http.StatusUnauthorized, "Missing authorization token")
	}
	return "", false
}

func getAccessToken(c *gin.Context) (string, bool) {
	authHeader := c.Request.Header.Get("Authorization")
	log.Printf("Auth Header: %s\n", authHeader)
	if authHeader != "" {
		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) == 2 {
			reqToken := splitToken[1]
			return reqToken, true
		}
	}

	return "", false
}

func verifyAccessToken(accessToken string) (*oauth2.Tokeninfo, error) {
	oauth2Service, err := oauth2.New(httpClient)
	tokenInfoCall := oauth2Service.Tokeninfo()
	tokenInfoCall.AccessToken(accessToken)
	tokenInfo, err := tokenInfoCall.Do()
	if err != nil {
		return nil, err
	}
	return tokenInfo, nil
}
