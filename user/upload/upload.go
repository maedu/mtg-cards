package upload

import (
	"encoding/csv"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/dimchansky/utfbom"
	"github.com/gin-gonic/gin"
	"github.com/maedu/mtg-cards/user/auth"
	"github.com/maedu/mtg-cards/user/db"
)

func Setup(r *gin.Engine) {
	r.POST("/api/user/cards/upload/:source", handleUploadCards)
}

func handleUploadCards(c *gin.Context) {
	source := c.Param("source")
	if source == "" {
		c.JSON(http.StatusBadRequest, "Parameter source is missing")
		return
	}

	if userID, ok := auth.GetUserIDFromAccessToken(c, true); ok {
		uploadCards(c, ParseRequest{UserID: userID, Source: source})
	}
}

func uploadCards(c *gin.Context, request ParseRequest) {

	parsedCards, err := parseUserCards(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	collection, err := db.GetUserCardCollection()
	defer collection.Disconnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	existingCards, err := collection.GetUserCardsByUserID(request.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	// Create a map for easier updating
	cardMap := map[string]*db.UserCard{}
	for _, card := range existingCards {
		cardMap[card.Name] = card
	}

	// Merge existing and parsed cards
	var found bool
	for _, parsedCard := range parsedCards {
		set := db.Set{
			ID:       parsedCard.Set,
			Quantity: parsedCard.Quantity,
			Source:   request.Source,
		}

		if cardMap[parsedCard.Name] != nil {
			found = false
			for _, set := range cardMap[parsedCard.Name].Sets {

				if set.ID == parsedCard.Set {
					set.Quantity = parsedCard.Quantity
					found = true
					break
				}
			}
			if !found {
				cardMap[parsedCard.Name].Sets = append(cardMap[parsedCard.Name].Sets, set)
			}
		} else {
			card := db.UserCard{
				Name:   parsedCard.Name,
				UserID: request.UserID,
				Sets:   []db.Set{set},
			}
			cardMap[parsedCard.Name] = &card
		}
	}

	cards := []*db.UserCard{}
	for _, card := range cardMap {
		cards = append(cards, card)
	}

	err = collection.ReplaceAllOfUser(request.UserID, cards)
	if err != nil {
		fmt.Printf("Error while storing: %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, cards)
}

func readFile(c *gin.Context) (*multipart.File, error) {
	file, _, err := c.Request.FormFile("file")
	return &file, err
}

func getCsvReader(content *multipart.File) *csv.Reader {
	sr, enc := utfbom.Skip(*content)
	fmt.Printf("CSV: Detected encoding: %s\n", enc)
	r := csv.NewReader(sr)
	r.Comma = ','
	r.FieldsPerRecord = -1 // Enable variable number of fields
	r.LazyQuotes = true
	return r
}

func parseUserCards(c *gin.Context) ([]*ParsedCard, error) {

	handlers := []handler{mtgGoldFish{}}

	for _, handler := range handlers {
		content, err := readFile(c)

		if err != nil {
			return nil, err
		}

		csvReader := getCsvReader(content)
		success, parsedCards, err := handler.parse(csvReader)
		if err != nil {
			fmt.Println(err)
		}
		if success {
			fmt.Printf("Upload done with handler %s", handler.name())
			return parsedCards, nil
		}
	}

	return nil, errors.New("no parser matched")
}
