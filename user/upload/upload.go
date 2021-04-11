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
		uploadCards(c, userID, source)
	}
}

func uploadCards(c *gin.Context, user string, source string) {

	cards, err := parseUserCards(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	for _, card := range cards {
		card.UserID = user
		card.Source = source
	}
	collection, err := db.GetUserCardCollection()
	defer collection.Disconnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	err = collection.ReplaceAllOfUserAndSource(user, source, cards)
	if err != nil {
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

func parseUserCards(c *gin.Context) ([]*db.UserCard, error) {

	handlers := []handler{mtgGoldFish{}}

	for _, handler := range handlers {
		content, err := readFile(c)

		if err != nil {
			return nil, err
		}

		csvReader := getCsvReader(content)
		success, cards, err := handler.parse(csvReader)
		if err != nil {
			fmt.Println(err)
		}
		if success {
			fmt.Printf("Upload done with handler %s", handler.name())
			return cards, nil
		}
	}

	return nil, errors.New("no parser matched")
}
