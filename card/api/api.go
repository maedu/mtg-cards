package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/maedu/mtg-cards/card/db"
)

type FindCardsRequest struct {
	Cards []string `json:"cards"`
	Sets  []string `json:"sets"`
}
type FindCardsResponse struct {
	Cards map[string]*db.Card   `json:"cards"`
	Sets  map[string][]*db.Card `json:"sets"`
}

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/cards", handleGetCards)
	r.GET("/api/cards/set", handleGetSet)
	r.POST("/api/cards/find", handleFindCards)
	r.GET("/api/cards/update", handleUpdateCards)
	r.GET("/api/cards/transform", handleTransformCards)

}

func handleGetSet(c *gin.Context) {

	var err error

	set := c.Query("set")
	collection, err := db.GetCardCollection()
	if err != nil {
		c.Error(err)
		return
	}
	defer collection.Disconnect()

	loadedCards, err := collection.GetCardsBySetName(set)
	if err != nil {
		c.Error(err)
		return
	}

	names := ""
	for _, card := range loadedCards {
		names = names + "\n1 " + card.Name
	}

	c.JSON(http.StatusOK, names)
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

	collection, err := db.GetCardCollection()
	if err != nil {
		c.Error(err)
		return
	}
	defer collection.Disconnect()

	text := c.Query("text")
	cmcText := c.QueryArray("cmc")
	cmc := []float64{}
	if cmcText != nil {
		for _, item := range cmcText {
			result, err := strconv.ParseFloat(item, 0)
			if err != nil {
				c.Error(err)
				return
			}
			cmc = append(cmc, result)
		}
	}
	colors := c.QueryArray("colors")
	cardGroups := c.QueryArray("cardGroups")

	mainCardForSynergy := c.Query("mainCardForSynergy")
	searchRelatedToMainCard := c.Query("searchRelatedToMainCard") == "true"

	priceMinString := c.Query("priceMin")
	priceMin := db.PriceFilterSkipped
	if priceMinString != "" {
		priceMin, err = strconv.ParseFloat(priceMinString, 0)
		if err != nil {
			c.Error(err)
			return
		}
	}
	priceMaxString := c.Query("priceMax")
	priceMax := db.PriceFilterSkipped
	if priceMaxString != "" {
		priceMax, err = strconv.ParseFloat(priceMaxString, 0)
		if err != nil {
			c.Error(err)
			return
		}
	}

	sortBy := c.Query("sortBy")
	sortDir := c.Query("sortDir")

	request := db.CardSearchRequest{
		Text:                    text,
		Cmc:                     cmc,
		Colors:                  colors,
		CardGroups:              cardGroups,
		MainCardForSynergy:      mainCardForSynergy,
		SearchRelatedToMainCard: searchRelatedToMainCard,
		PriceMin:                priceMin,
		PriceMax:                priceMax,
		SortBy:                  sortBy,
		SortDir:                 sortDir,
	}
	loadedCards, err := collection.GetCardsPaginated(perPage, page, request)
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
		c.JSON(http.StatusBadRequest, err)
		return
	}
	collection, err := db.GetCardCollection()
	if err != nil {
		c.Error(err)
		return
	}
	defer collection.Disconnect()

	cards := map[string]*db.Card{}

	cardsToFind := make([]string, len(request.Cards))
	copy(cardsToFind, request.Cards)

	for _, name := range request.Cards {
		cards[name] = nil
	}
	foundCards, err := collection.GetCardsByNames(cardsToFind)
	if err != nil {
		c.Error(err)
		return
	}

	for _, card := range foundCards {
		cards[card.Name] = card
	}

	sets := map[string][]*db.Card{}
	for _, name := range request.Sets {
		cards, err := collection.GetCardsBySetName(name)
		if err != nil {
			c.Error(err)
			return
		}
		sets[name] = cards
	}

	response := FindCardsResponse{
		Cards: cards,
		Sets:  sets,
	}

	c.JSON(http.StatusOK, response)
}
