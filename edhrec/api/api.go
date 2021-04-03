package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maedu/mtg-cards/card/db"
	edhrecDB "github.com/maedu/mtg-cards/edhrec/db"
	"github.com/maedu/mtg-cards/edhrec/parser"
)

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/edhrec/commander/:name", handleCommander)
	r.GET("/api/edhrec/synergy/:name", handleSynergy)
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
		cards[edhRecCard.CardWithSynergy] = nil
		cardNames = append(cardNames, edhRecCard.CardWithSynergy)
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

func handleSynergy(c *gin.Context) {
	mainCard := c.Param("name")
	edhRecCards, err := parser.FetchCommander(mainCard)
	if err != nil {
		c.Error(err)
		return
	}
	err = storeSynergyInDb(mainCard, edhRecCards)
	if err != nil {
		c.Error(err)
		return
	}

	cards := map[string]float64{}
	for _, edhRecCard := range edhRecCards {
		cards[edhRecCard.CardWithSynergy] = edhRecCard.Synergy
	}
	c.JSON(http.StatusOK, cards)
}

func storeSynergyInDb(mainCard string, edhRecCards []edhrecDB.EdhrecSynergy) error {
	collection := edhrecDB.GetEdhrecSynergyCollection()
	defer collection.Disconnect()
	return collection.ReplaceAllOfMainCard(mainCard, edhRecCards)
}
