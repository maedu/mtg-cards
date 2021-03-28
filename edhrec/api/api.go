package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maedu/mtg-cards/card/db"
	"github.com/maedu/mtg-cards/edhrec/parser"
)

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/edhrec/commander/:name", handleCommander)
}

func handleCommander(c *gin.Context) {
	name := c.Param("name")
	edhRecCards, err := parser.FetchCommander(name)
	if err != nil {
		c.Error(err)
		return
	}

	collection, err := db.GetCardCollection()
	if err != nil {
		c.Error(err)
		return
	}
	defer collection.Disconnect()

	cards := map[string]*db.Card{}
	cardNames := []string{}
	for _, edhRecCard := range edhRecCards {
		cards[edhRecCard.Name] = nil
		cardNames = append(cardNames, edhRecCard.Name)
	}
	foundCards, err := collection.GetCardsByNames(cardNames)
	if err != nil {
		c.Error(err)
		return
	}

	for _, card := range foundCards {
		cards[card.Name] = card
	}
	c.JSON(http.StatusOK, cards)
}
