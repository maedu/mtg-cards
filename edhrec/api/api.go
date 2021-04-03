package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	cardAPI "github.com/maedu/mtg-cards/card/api"
	edhrecDB "github.com/maedu/mtg-cards/edhrec/db"
	"github.com/maedu/mtg-cards/edhrec/parser"
)

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/edhrec/synergy/:name", handleSynergy)
}

func handleSynergy(c *gin.Context) {
	mainCard := c.Param("name")
	update := c.Query("update")

	collection, err := edhrecDB.GetEdhrecSynergyCollection()
	defer collection.Disconnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	edhRecCards, err := collection.GetEdhrecSynergysByMainCard(mainCard)
	collection.Disconnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	if len(edhRecCards) == 0 || update != "" {
		edhRecCards, err := parser.FetchCommander(mainCard)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		collection, err = edhrecDB.GetEdhrecSynergyCollection()
		err = collection.ReplaceAllOfMainCard(mainCard, edhRecCards)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		err = cardAPI.TransformCards()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

	}

	cards := map[string]float64{}
	for _, edhRecCard := range edhRecCards {
		cards[edhRecCard.CardWithSynergy] = edhRecCard.Synergy
	}
	c.JSON(http.StatusOK, cards)
}
