package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/maedu/mtg-cards/scryfall/db"
)

type SetsBulkDataResponse struct {
	Data []*db.ScryfallSet `json:"data"`
}

func UpdateSets() error {
	log.Println("UpdateSets")

	sets, err := getSets()
	if err != nil {
		return fmt.Errorf("getSets: %w", err)
	}
	log.Println("GetScryfallSetCollection")
	collection := db.GetScryfallSetCollection()
	defer collection.Disconnect()
	_, err = collection.ReplaceAll(sets)
	return err
}

func getSets() ([]*db.ScryfallSet, error) {
	url := "https://api.scryfall.com/sets"
	log.Printf("Get Sets: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		log.Printf("Status not 200 but %d", resp.StatusCode)
		return nil, fmt.Errorf("Status not 200 but %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	var res SetsBulkDataResponse
	err = json.NewDecoder(resp.Body).Decode(&res)

	if err != nil {
		log.Print(err)
		return nil, err
	}
	return res.Data, nil
}
