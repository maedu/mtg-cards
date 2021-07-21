package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	cardDB "github.com/maedu/mtg-cards/card/db"
	"github.com/maedu/mtg-cards/deck/db"
	"github.com/maedu/mtg-cards/user/auth"
	userDB "github.com/maedu/mtg-cards/user/db"
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
	r.POST("/api/decks/:urlHash/publish", handlePublishDeck)
	r.POST("/api/decks/:urlHash/unpublish", handleUnpublishDeck)
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
	if deck == nil {
		c.JSON(http.StatusNotFound, "Deck not found")
		return
	}
	userID, _ := auth.GetUserIDFromAccessToken(c, false)
	if !canUserSeeDeck(deck, userID) {
		c.JSON(http.StatusForbidden, "You are not allowed to see this deck")
		return
	}

	deckWithCards, err := dbDeckToDeck(deck)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, deckWithCards)
}

func canUserSeeDeck(deck *db.Deck, userID string) bool {
	return deck.UserID == userID || deck.Settings.Published
}

func handleGetUserDecks(c *gin.Context) {
	selectedUserName := c.Query("user")
	var selectedUserID string
	userID, loggedIn := auth.GetUserIDFromAccessToken(c, false)
	if selectedUserName != "" {
		userCollection, err := userDB.GetUserCollection()
		defer userCollection.Disconnect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		user, err := userCollection.GetUserByUserName(selectedUserName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		if user == nil {
			c.JSON(http.StatusNotFound, "User not found")
			return
		}

		selectedUserID = user.UserID
	} else {
		if !loggedIn {
			c.JSON(http.StatusUnauthorized, nil)
			return
		}
		selectedUserID = userID
	}

	collection, err := db.GetDeckCollection()
	defer collection.Disconnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	publishedOnly := selectedUserID != userID
	decks, err := collection.GetDecksByUserID(selectedUserID, publishedOnly)
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

func handlePublishDeck(c *gin.Context) {
	setPublish(c, true)
}
func handleUnpublishDeck(c *gin.Context) {
	setPublish(c, false)
}

func setPublish(c *gin.Context, publish bool) {
	if userID, ok := auth.GetUserIDFromAccessToken(c, true); ok {
		urlHash := c.Param("urlHash")
		if urlHash == "" {
			c.JSON(http.StatusBadRequest, "Missing urLHash parameter")
			return
		}

		collection, err := db.GetDeckCollection()
		defer collection.Disconnect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		storedDeck, err := collection.GetDeckByURLHash(urlHash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		if storedDeck == nil {
			c.JSON(http.StatusNotFound, "Deck not found")
			return
		}

		if storedDeck.UserID != userID {
			c.JSON(http.StatusForbidden, err)
			return
		}

		storedDeck.Settings.Published = publish
		storedDeck, err = collection.Update(storedDeck)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, storedDeck)
		return

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
		if deckIsEmpty(&inputDeck) {
			c.JSON(http.StatusBadRequest, "No cards selected, a deck needs at least one card")
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

func deckIsEmpty(deck *Deck) bool {
	if len(deck.Commanders) > 0 {
		return false
	}
	if len(deck.Deck) > 0 {
		return false
	}

	return true
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
		Deck:       []*cardDB.Card{},
		Library:    []*cardDB.Card{},
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
