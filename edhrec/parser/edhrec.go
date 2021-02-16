package parser

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type CommandersJSON struct {
	Panels    Panels    `json:"panels"`
	Container Container `json:"container"`
}

type Panels struct {
	Card Card `json:"card"`
}

type Card struct {
	Name string `json:"name"`
}

type Container struct {
	JsonDict JsonDict `json:"json_dict"`
}

type JsonDict struct {
	CardLists []CardList `json:"cardlists"`
}

type CardList struct {
	CardViews []CardView `json:"cardviews"`
}

type CardView struct {
	Name string `json:"name"`
}

func FetchCommander(commander string) ([]string, error) {
	url := fmt.Sprintf("https://edhrec-json.s3.amazonaws.com/en/commanders/%s.json", commander)
	content, err := fetchUrl(url)
	if err != nil {
		return nil, err
	}

	cards := []string{content.Panels.Card.Name}

	for _, cardList := range content.Container.JsonDict.CardLists {
		for _, card := range cardList.CardViews {
			cards = append(cards, card.Name)
		}
	}

	return cards, nil
}

func fetchUrl(url string) (CommandersJSON, error) {

	log.Printf("Fetch URL: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Print(err)
		return CommandersJSON{}, fmt.Errorf("Fetching failed: %w", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("Status not 200 but %d", resp.StatusCode)
		return CommandersJSON{}, fmt.Errorf("Status not 200 but %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var res CommandersJSON
	err = json.NewDecoder(resp.Body).Decode(&res)

	if err != nil {
		log.Print(err)
		return CommandersJSON{}, err
	}
	return res, nil
}
