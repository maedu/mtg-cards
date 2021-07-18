package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maedu/mtg-cards/deck/db"
	"github.com/maedu/mtg-cards/user/auth"
	"github.com/maedu/mtg-cards/util"
)

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/deck/:urlHash", handlGetDeck)
	r.POST("/api/deck", handleUpsertDeck)
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

func handleUpsertDeck(c *gin.Context) {
	if userID, ok := auth.GetUserIDFromAccessToken(c, true); ok {
		var deck *db.Deck
		c.BindJSON(deck)
		collection, err := db.GetDeckCollection()
		defer collection.Disconnect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		var storedDeck *db.Deck
		if deck.URLHash == "" {
			// New deck
			hash, err := generateUniqueURLHash(&collection)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			deck.URLHash = hash
			deck.UserID = userID
			_, err = collection.Create(deck)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			storedDeck = deck
		} else {
			storedDeck, err = collection.GetDeckByURLHash(deck.URLHash)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			if storedDeck.UserID ö¨!.ä$$$$= userID {
				c.JSON(http.StatusForbidden, err)
				return
			}

			storedDeck, err = collection.Update(storedDeck)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}

		c.JSON(http.StatusOK, storedDeck)

	}
	c.JSON(http.StatusUnauthorized, nil)

}

func generateUniqueURLHash(collection *db.DeckCollection) (string, error) {
	hash := util.RandomString(10)
	for {
		checkedDeck, err := collection.GetDeckByURLHash(hash)
		if err != nil {
			return "", err
		}
		if checkedDeck == nil {
			break
		}
		hash = util.RandomString(20)
	}
	return hash, nil
}
