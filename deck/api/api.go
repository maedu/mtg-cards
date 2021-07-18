package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maedu/mtg-cards/deck/db"
)

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/deck/:urlHash", handlGetDeck)
}
func handlGetDeck(c *gin.Context) {
	urlHash := c.Param("urlHash")

	collection, err := db.GetDeckCollection()
	defer collection.Disconnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	deck, err := collection.GetDeckByURLHash(urlHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, deck)
}
