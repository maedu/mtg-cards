package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	cardDB "github.com/maedu/mtg-cards/card/db"
	"github.com/maedu/mtg-cards/deck/db"
	"github.com/maedu/mtg-cards/user/auth"
	"github.com/maedu/mtg-cards/util"
)

type Deck struct {
	Commanders []*cardDB.Card `json:"commanders"`
	Deck       []*cardDB.Card `json:"deck"`
	Library    []*cardDB.Card `json:"library"`
	Settings   db.Settings    `json:"settings"`
}

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/decks/:urlHash", handlGetDeck)
	r.POST("/api/decks", handleUpsertDeck)
	r.GET("/api/decks", handleGetUserDecks)
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
	deckWithCards, err := dbDeckToDeck(deck)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, deckWithCards)
}

func handleGetUserDecks(c *gin.Context) {
	if userID, ok := auth.GetUserIDFromAccessToken(c, true); ok {

		collection, err := db.GetDeckCollection()
		defer collection.Disconnect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		decks, err := collection.GetDecksByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		decksWithCards := []Deck{}
		for _, deck := range decks {
			deckWithCards, err := dbDeckToDeckForOverview(deck)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			decksWithCards = append(decksWithCards, deckWithCards)
		}

		c.JSON(http.StatusOK, decksWithCards)

	}

	c.JSON(http.StatusUnauthorized, nil)
}

func handleUpsertDeck(c *gin.Context) {
	if userID, ok := auth.GetUserIDFromAccessToken(c, true); ok {
		var inputDeck Deck
		err := c.BindJSON(&inputDeck)
		if err != nil {
			fmt.Printf("%+v", err)
			// c.BindJSON already sets the header to 400
			return
		}

		deck := deckToDBDeck(&inputDeck)
		deck.UserID = userID

		collection, err := db.GetDeckCollection()
		defer collection.Disconnect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		var storedDeck *db.Deck
		if deck.Settings.URLHash == "" {
			// New deck
			fmt.Println("New Deck")
			hash, err := generateUniqueURLHash(&collection)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			deck.Settings.URLHash = hash
			_, err = collection.Create(&deck)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			storedDeck = &deck
		} else {
			fmt.Printf("Update Deck: url = %s\n", deck.Settings.URLHash)
			storedDeck, err = collection.GetDeckByURLHash(deck.Settings.URLHash)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			if storedDeck != nil {
				fmt.Printf("Stored deck found, %s == %s\n", storedDeck.UserID, userID)
				if storedDeck.UserID != userID {
					c.JSON(http.StatusForbidden, err)
					return
				}
				deck.ID = storedDeck.ID
				storedDeck, err = collection.Update(&deck)
				if err != nil {
					c.JSON(http.StatusInternalServerError, err)
					return
				}
			} else {
				// Special case: Deck does not exist in DB
				fmt.Println("Special case")
				_, err = collection.Create(&deck)
				if err != nil {
					c.JSON(http.StatusInternalServerError, err)
					return
				}
				storedDeck = &deck
			}
		}

		c.JSON(http.StatusOK, storedDeck.Settings.URLHash)
		return

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

func deckToDBDeck(deck *Deck) db.Deck {
	return db.Deck{
		Commanders: cardListToNames(deck.Commanders),
		Deck:       cardListToNames(deck.Deck),
		Library:    cardListToNames(deck.Library),
		Settings:   deck.Settings,
	}
}

func cardListToNames(cards []*cardDB.Card) []string {
	names := []string{}
	for _, card := range cards {
		names = append(names, card.Name)
	}
	return names
}

func dbDeckToDeckForOverview(deck *db.Deck) (Deck, error) {
	collection, err := cardDB.GetCardCollection()
	defer collection.Disconnect()
	if err != nil {
		return Deck{}, err
	}
	commanders, err := collection.GetCardsByNames(deck.Commanders)
	if err != nil {
		return Deck{}, err
	}
	return Deck{
		Commanders: commanders,
		Settings:   deck.Settings,
	}, nil
}

func dbDeckToDeck(deck *db.Deck) (Deck, error) {
	collection, err := cardDB.GetCardCollection()
	defer collection.Disconnect()
	if err != nil {
		return Deck{}, err
	}
	commanders, err := collection.GetCardsByNames(deck.Commanders)
	if err != nil {
		return Deck{}, err
	}
	deckCards, err := collection.GetCardsByNames(deck.Deck)
	if err != nil {
		return Deck{}, err
	}
	library, err := collection.GetCardsByNames(deck.Library)
	if err != nil {
		return Deck{}, err
	}

	return Deck{
		Commanders: commanders,
		Deck:       deckCards,
		Library:    library,
		Settings:   deck.Settings,
	}, nil
}
