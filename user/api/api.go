package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maedu/mtg-cards/user/auth"
	userCardDB "github.com/maedu/mtg-cards/user/db"
)

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/user", handleGetUser)
	r.GET("/api/user/cards", handleGetCards)
}

func handleGetUser(c *gin.Context) {
	if userID, ok := auth.GetUserIDFromAccessToken(c, true); ok {
		c.JSON(http.StatusOK, userID)
	}
}

func handleGetCards(c *gin.Context) {
	if userID, ok := auth.GetUserIDFromAccessToken(c, true); ok {
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
