package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/maedu/mtg-cards/scryfall/db"
)

const oracleCards = "oracle_cards"

type BulkDataResponse struct {
	Data []*BulkData `json:"data"`
}

type BulkData struct {
	Type        string `json:"type"`
	DownloadURI string `json:"download_uri"`
}

func UpdateCards() error {
	log.Println("UpdateCards")
	bulkData, err := getBulkData()
	if err != nil {
		return fmt.Errorf("getBulkData: %w", err)
	}

	uri, err := uri(bulkData)
	if err != nil {
		return fmt.Errorf("uri: %w", err)
	}

	cards, err := getCards(uri)
	if err != nil {
		return fmt.Errorf("getCards: %w", err)
	}
	log.Println("GetScryfallCardCollection")
	collection := db.GetScryfallCardCollection()
	defer collection.Disconnect()
	_, err = collection.ReplaceAll(cards)
	return err
}

func getBulkData() ([]*BulkData, error) {
	url := "https://api.scryfall.com/bulk-data"
	log.Printf("Get BulkData: %s", url)

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
	var res BulkDataResponse
	err = json.NewDecoder(resp.Body).Decode(&res)

	if err != nil {
		log.Print(err)
		return nil, err
	}
	return res.Data, nil
}

func uri(bulkDataSlice []*BulkData) (string, error) {
	for _, bulkData := range bulkDataSlice {
		if bulkData.Type == oracleCards {
			return bulkData.DownloadURI, nil
		}
	}

	return "", fmt.Errorf("no bulk_data for oracleCards found")
}

func getCards(uri string) ([]*db.ScryfallCard, error) {
	log.Printf("Get Cards: %s", uri)
	resp, err := http.Get(uri)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		log.Printf("Status not 200 but %d", resp.StatusCode)
		return nil, fmt.Errorf("Status not 200 but %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	var res []*db.ScryfallCard
	err = json.NewDecoder(resp.Body).Decode(&res)

	if err != nil {
		log.Print(err)
		return nil, err
	}
	return res, nil
}
