package parser

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type CommandersJSON struct {
	Container Container `json:"container"`
}

type Container struct {
	JsonDict JsonDict `json:"json_dict"`
}

type JsonDict struct {
	Card      Card       `json:"card"`
	CardLists []CardList `json:"cardlists"`
}

type Card struct {
	Name string `json:"name"`
}

type CardList struct {
	CardViews []CardView `json:"cardviews"`
}

type CardView struct {
	Name    string  `json:"name"`
	Synergy float64 `json:"synergy"`
}

type EdhRecCard struct {
	Name    string  `json:"name"`
	Synergy float64 `json:"synergy"`
}

func FetchCommander(commander string) ([]EdhRecCard, error) {
	routeUrl, err := getRoute(commander)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://edhrec-json.s3.amazonaws.com/en%s.json", routeUrl)
	content, err := fetchUrl(url)
	if err != nil {
		return nil, err
	}

	cards := []EdhRecCard{
		EdhRecCard{
			Name:    content.Container.JsonDict.Card.Name,
			Synergy: 0,
		},
	}

	log.Printf("Commander name: %s", content.Container.JsonDict.Card.Name)

	for _, cardList := range content.Container.JsonDict.CardLists {
		for _, card := range cardList.CardViews {
			cards = append(cards, EdhRecCard{
				Name:    card.Name,
				Synergy: card.Synergy,
			})
		}
	}

	return cards, nil
}

func getRoute(commander string) (string, error) {
	routeUrl := fmt.Sprintf("https://edhrec.com/api/route/?cc=%s", strings.Replace(commander, " ", "%20", -1))

	log.Printf("Get Route URL: %s", routeUrl)

	resp, err := http.Get(routeUrl)
	if err != nil {
		log.Print(err)
		return "", fmt.Errorf("Fetching failed: %w", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("Status not 200 but %d", resp.StatusCode)
		return "", fmt.Errorf("Status not 200 but %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var res string
	err = json.NewDecoder(resp.Body).Decode(&res)

	if err != nil {
		log.Print(err)
		return "", err
	}
	log.Printf("Route URL result: %s", res)
	return res, nil
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
