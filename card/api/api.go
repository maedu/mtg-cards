package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"bitbucket.org/spinnerweb/cards/card/db"
	"bitbucket.org/spinnerweb/cards/scryfall/client"
	scryfallDB "bitbucket.org/spinnerweb/cards/scryfall/db"
	"github.com/gin-gonic/gin"
)

type FindCardsRequest struct {
	Names []string `json:"names"`
}

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/cards", handleGetCards)
	r.POST("/api/cards/find", handleFindCards)
	r.GET("/api/cards/update", handleUpdateCards)
	r.GET("/api/cards/transform", handleTransformCards)

}

func handleGetCards(c *gin.Context) {

	var err error

	pageS := c.Query("page")
	var page int64 = 1
	if pageS != "" {
		page, err = parseInt(pageS)
		if err != nil {
			c.Error(err)
			return
		}
	}
	perPageS := c.Query("perPage")
	var perPage int64 = 100
	if perPageS != "" {
		perPage, err = parseInt(perPageS)
		if err != nil {
			c.Error(err)
			return
		}
	}

	textSearch := c.Query("textSearch")

	collection, err := db.GetCardCollection()
	if err != nil {
		c.Error(err)
		return
	}
	defer collection.Disconnect()

	loadedCards, err := collection.GetCardsPaginated(perPage, page, textSearch)
	//loadedCards, err := collection.FindCards(filterByFullText)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, loadedCards)
}

func handleFindCards(c *gin.Context) {

	var request FindCardsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Println(err)
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"msg": err})
		return
	}
	collection, err := db.GetCardCollection()
	if err != nil {
		c.Error(err)
		return
	}
	defer collection.Disconnect()

	response := map[string]*db.Card{}

	for _, name := range request.Names {
		card, err := collection.GetCardByName(name)
		if err != nil {
			c.Error(err)
			return
		}
		response[name] = card
	}

	c.JSON(http.StatusOK, response)
}

func handleUpdateCards(c *gin.Context) {

	err := client.UpdateCards()
	if err != nil {
		c.Error(err)
		return
	}

	err = transformCards()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, nil)

}

func handleTransformCards(c *gin.Context) {

	err := transformCards()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, nil)

}

func transformCards() error {

	log.Println("Get scryfallCollection")
	scryfallCollection := scryfallDB.GetScryfallCardCollection()
	defer scryfallCollection.Disconnect()

	var loadedScryfallCards, err = scryfallCollection.GetAllScryfallCards()

	if err != nil {
		return err
	}

	cards := []*db.Card{}
	log.Println("Transform cards")
	for _, scryfallCard := range loadedScryfallCards {
		card := transformCard(scryfallCard)
		if card != nil {
			cards = append(cards, card)
		}
	}

	log.Println("Get cards collection")
	collection, err := db.GetCardCollection()
	if err != nil {
		return err
	}
	defer collection.Disconnect()
	_, err = collection.ReplaceAll(cards)

	return err
}

func transformCard(scryfallCard *scryfallDB.ScryfallCard) *db.Card {

	switch scryfallCard.Layout {
	case "art_series":
	case "token":
	case "emblem":
		// Ignore those types
		return nil
	}

	legalInCommander := false
	if len(scryfallCard.Legalities) > 0 {
		for gameType, legalText := range scryfallCard.Legalities {
			if gameType == scryfallDB.Commander {
				legalInCommander = legalText == scryfallDB.Legal
				if legalInCommander {
					break
				}
			}
		}
	}

	imageURLs := map[string]string{}

	for size, url := range scryfallCard.ImageURLs {
		if size == scryfallDB.Large || size == scryfallDB.Normal {
			imageURLs[size] = url
		}
	}

	price := 0.0
	for currency, currencyPrice := range scryfallCard.Prices {
		if currency == scryfallDB.USD {
			price = parseAmount(currencyPrice)
		}
	}

	cardFaces := []db.Card{}
	for _, scryfallCardFace := range scryfallCard.CardFaces {
		cardFace := transformCard(&scryfallCardFace)
		if cardFace != nil {
			cardFaces = append(cardFaces, *cardFace)
		}
	}

	isCommander := strings.Contains(scryfallCard.TypeLine, "Legendary Creature") || strings.Contains(scryfallCard.OracleText, "can be your commander")

	rarity := scryfallCard.Rarity
	if rarity == "mythic" {
		rarity = "mythic rare"
	}
	searchText := strings.ToLower(fmt.Sprintf("%s, %s, %s, %s, %v, %s",
		scryfallCard.Name,
		scryfallCard.ManaCost,
		scryfallCard.TypeLine,
		scryfallCard.OracleText,
		scryfallCard.Keywords,
		rarity,
	))

	cardTypesToCheck := []db.CardType{
		db.Artifact,
		db.Creature,
		db.Enchantment,
		db.Instant,
		db.Land,
		db.Plane,
		db.Planeswalker,
		db.Sorcery,
		db.Conspiracy,
		db.Phenomenon,
		db.Scheme,
		db.Tribal,
		db.Vanguard,
	}

	var cardType db.CardType

	for _, cardTypeToCheck := range cardTypesToCheck {
		if strings.Contains(scryfallCard.TypeLine, string(cardTypeToCheck)) {
			cardType = cardTypeToCheck
			break
		}
	}

	card := db.Card{
		ScryfallID:      scryfallCard.ID,
		Name:            scryfallCard.Name,
		Lang:            scryfallCard.Lang,
		ImageURLs:       imageURLs,
		ManaCost:        scryfallCard.ManaCost,
		Cmc:             scryfallCard.Cmc,
		TypeLine:        scryfallCard.TypeLine,
		CardType:        cardType,
		OracleText:      scryfallCard.OracleText,
		Colors:          scryfallCard.Colors,
		ColorIdentity:   scryfallCard.ColorIdentity,
		Keywords:        scryfallCard.Keywords,
		LegalInComander: legalInCommander,
		SetName:         scryfallCard.SetName,
		RulingsURL:      scryfallCard.RulingsURI,
		Rarity:          scryfallCard.Rarity,
		EdhrecRank:      scryfallCard.EdhrecRank,
		Layout:          scryfallCard.Layout,
		Price:           price,
		CardFaces:       cardFaces,
		IsCommander:     isCommander,
		SearchText:      searchText,
	}

	return &card
}

func parseAmount(amount string) float64 {
	val := strings.ReplaceAll(amount, "'", "")
	val = strings.ReplaceAll(val, ",", "")
	val = strings.TrimSpace(val)
	f, _ := strconv.ParseFloat(val, 64)
	return f
}

func parseInt(number string) (int64, error) {
	val := strings.ReplaceAll(number, "'", "")
	val = strings.ReplaceAll(val, ",", "")
	val = strings.TrimSpace(val)
	return strconv.ParseInt(val, 10, 64)
}
