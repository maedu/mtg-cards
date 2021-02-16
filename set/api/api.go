package api

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maedu/mtg-cards/scryfall/client"
	scryfallDB "github.com/maedu/mtg-cards/scryfall/db"
	"github.com/maedu/mtg-cards/set/db"
)

// Setup Setup REST API
func Setup(r *gin.Engine) {
	r.GET("/api/sets", handleGetSets)
	r.GET("/api/sets/update", handleUpdateSets)
	r.GET("/api/sets/transform", handleTransformSets)

}

func handleGetSet(c *gin.Context) {

	var err error

	set := c.Query("set")
	collection, err := db.GetSetCollection()
	if err != nil {
		c.Error(err)
		return
	}
	defer collection.Disconnect()

	loadedSets, err := collection.GetSetsBySetName(set)
	if err != nil {
		c.Error(err)
		return
	}

	names := ""
	for _, set := range loadedSets {
		names = names + "\n1 " + set.Name
	}

	c.JSON(http.StatusOK, names)
}

func handleGetSets(c *gin.Context) {

	var err error

	collection, err := db.GetSetCollection()
	if err != nil {
		c.Error(err)
		return
	}
	defer collection.Disconnect()

	loadedSets, err := collection.GetAllSets()
	//loadedSets, err := collection.FindSets(filterByFullText)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, loadedSets)
}

func handleUpdateSets(c *gin.Context) {

	err := client.UpdateSets()
	if err != nil {
		c.Error(err)
		return
	}

	err = transformSets()
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, nil)

}

func handleTransformSets(c *gin.Context) {

	err := transformSets()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, nil)

}

func transformSets() error {

	log.Println("Get scryfallCollection")
	scryfallCollection := scryfallDB.GetScryfallSetCollection()
	defer scryfallCollection.Disconnect()

	var loadedScryfallSets, err = scryfallCollection.GetAllScryfallSets()

	if err != nil {
		return err
	}

	sets := []*db.Set{}
	log.Println("Transform sets")
	for _, scryfallSet := range loadedScryfallSets {
		set, err := transformSet(scryfallSet)
		if err != nil {
			return err
		}
		if set != nil {
			sets = append(sets, set)
		}
	}

	log.Println("Get sets collection")
	collection, err := db.GetSetCollection()
	if err != nil {
		return err
	}
	defer collection.Disconnect()
	_, err = collection.ReplaceAll(sets)

	return err
}

func transformSet(scryfallSet *scryfallDB.ScryfallSet) (*db.Set, error) {

	log.Printf("Transforming '%s', setType: '%s'", scryfallSet.Name, scryfallSet.SetType)
	switch scryfallSet.SetType {
	case "token", "memorabilia", "promo":
		// Ignore those types
		log.Printf("Set '%s' has wrong type '%s'", scryfallSet.Name, scryfallSet.SetType)
		return nil, nil
	}

	if scryfallSet.CardCount < 100 {
		log.Printf("Set '%s' has too few cards (%d < 100)", scryfallSet.Name, scryfallSet.CardCount)
		return nil, nil
	}

	releasedAt, err := time.Parse("2006-01-02", scryfallSet.ReleasedAt)
	if err != nil {
		return nil, err
	}

	set := &db.Set{
		ScryfallID: scryfallSet.ID,
		Name:       scryfallSet.Name,
		Code:       scryfallSet.Code,
		SetType:    scryfallSet.SetType,
		CardCount:  scryfallSet.CardCount,
		IconURL:    scryfallSet.IconSvgURI,
		ReleasedAt: releasedAt,
	}

	return set, nil
}

func parseInt(number string) (int64, error) {
	val := strings.ReplaceAll(number, "'", "")
	val = strings.ReplaceAll(val, ",", "")
	val = strings.TrimSpace(val)
	return strconv.ParseInt(val, 10, 64)
}
