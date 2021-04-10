package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	userCardDB "github.com/maedu/mtg-cards/user/db"
	oauth2 "google.golang.org/api/oauth2/v1"
)

var httpClient = &http.Client{}

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/user", handleGetUser)
	r.GET("/api/user/cards", handleGetCards)
}

func handleGetUser(c *gin.Context) {
	if userID, ok := getUserIDFromAccessToken(c); ok {
		c.JSON(http.StatusOK, userID)
	}
}

func handleGetCards(c *gin.Context) {
	if userID, ok := getUserIDFromAccessToken(c); ok {
		collection, err := userCardDB.GetUserCardCollection()
		defer collection.Disconnect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		userCards, err := collection.GetUserCardsByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, userCards)
		return
	}
}

func handleUploadCollection(c *gin.Context) {
	if userID, ok := getUserIDFromAccessToken(c); ok {
		collection, err := userCardDB.GetUserCardCollection()
		defer collection.Disconnect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		userCards, err := collection.GetUserCardsByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, userCards)
		return
	}
}

func getUserIDFromAccessToken(c *gin.Context) (string, bool) {

	if token, ok := getAccessToken(c); ok {
		info, err := verifyAccessToken(token)
		if err != nil {

			c.JSON(http.StatusForbidden, "Token verification failed")
			return "", false
		}

		return info.Email, true

	}
	c.JSON(http.StatusUnauthorized, "Missing authorization token")
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
