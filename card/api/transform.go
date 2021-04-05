package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/maedu/mtg-cards/card/cardgroup"
	"github.com/maedu/mtg-cards/card/db"
	edhrecDB "github.com/maedu/mtg-cards/edhrec/db"
	"github.com/maedu/mtg-cards/scryfall/client"
	scryfallDB "github.com/maedu/mtg-cards/scryfall/db"
)

func handleUpdateCards(c *gin.Context) {

	err := client.UpdateCards()
	if err != nil {
		c.Error(err)
		return
	}

	err = TransformCards()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, nil)

}

func handleTransformCards(c *gin.Context) {

	err := TransformCards()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, nil)

}

func TransformCards() error {

	log.Println("Get scryfallCollection")
	scryfallCollection := scryfallDB.GetScryfallCardCollection()
	defer scryfallCollection.Disconnect()

	loadedScryfallCards, err := scryfallCollection.GetAllScryfallCards()
	if err != nil {
		return err
	}

	edhrecSynergyCollection, err := edhrecDB.GetEdhrecSynergyCollection()
	defer edhrecSynergyCollection.Disconnect()

	if err != nil {
		return err
	}
	edhSynergies, err := edhrecSynergyCollection.GetAllEdhrecSynergys()
	synergies := map[string]map[string]float64{}
	for _, edhSynergy := range edhSynergies {
		if synergies[edhSynergy.CardWithSynergy] == nil {
			synergies[edhSynergy.CardWithSynergy] = map[string]float64{}
		}
		synergies[edhSynergy.CardWithSynergy][edhSynergy.MainCard] = edhSynergy.Synergy
	}

	cards := []*db.Card{}
	priceMap := getPriceMap(loadedScryfallCards)
	log.Println("Transform cards")
	for _, scryfallCard := range loadedScryfallCards {
		card := transformCard(scryfallCard, &synergies, &priceMap, nil)
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

func getPriceMap(loadedScryfallCards []*scryfallDB.ScryfallCard) map[scryfallDB.Currency]float64 {
	highestPrices := map[scryfallDB.Currency]float64{}
	var missing bool
	for _, scryfallCard := range loadedScryfallCards {
		prices := map[scryfallDB.Currency]float64{}
		missing = false
		for currency, currencyPrice := range scryfallCard.Prices {
			prices[currency] = parseAmount(currencyPrice)
			if prices[currency] == 0.0 {
				missing = true
				break
			}
		}
		if !missing {
			if highestPrices == nil || prices[scryfallDB.USD] > highestPrices[scryfallDB.USD] {
				highestPrices = prices
			}
		}
	}
	priceMap := map[scryfallDB.Currency]float64{}
	priceMap[scryfallDB.USD] = 1.0
	priceMap[scryfallDB.EUR] = highestPrices[scryfallDB.USD] / highestPrices[scryfallDB.EUR]
	priceMap[scryfallDB.TIX] = highestPrices[scryfallDB.USD] / highestPrices[scryfallDB.TIX]
	fmt.Printf("pricemAP: %v", priceMap)

	return priceMap
}

func transformCard(scryfallCard *scryfallDB.ScryfallCard, synergies *map[string]map[string]float64, priceMap *map[scryfallDB.Currency]float64, parentCard *scryfallDB.ScryfallCard) *db.Card {

	switch scryfallCard.Layout {
	case "art_series", "token", "double_faced_token", "emblem":
		// Ignore those types
		return nil
	}

	cardTypesToCheck := []db.CardType{
		db.Creature,
		db.Artifact,
		db.Enchantment,
		db.Instant,
		db.Land,
		db.Planeswalker,
		db.Sorcery,
		//db.Plane,
		//db.Conspiracy,
		//db.Phenomenon,
		//db.Scheme,
		//db.Tribal,
		//db.Vanguard,
	}

	cardTypes := []db.CardType{}
	isLand := false
	for _, cardTypeToCheck := range cardTypesToCheck {
		if strings.Contains(scryfallCard.TypeLine, string(cardTypeToCheck)) {
			cardTypes = append(cardTypes, cardTypeToCheck)
			if cardTypeToCheck == db.Land {
				isLand = true
			}
		}
	}
	if len(cardTypes) == 0 {
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

	cardFaces := []db.Card{}
	for _, scryfallCardFace := range scryfallCard.CardFaces {
		cardFace := transformCard(&scryfallCardFace, synergies, priceMap, scryfallCard)
		if cardFace != nil {
			cardFaces = append(cardFaces, *cardFace)
		}
	}

	isCommander := strings.Contains(scryfallCard.TypeLine, "Legendary Creature") || strings.Contains(scryfallCard.TypeLine, "Legendary Snow Creature") || strings.Contains(scryfallCard.OracleText, "can be your commander")

	rarity := scryfallCard.Rarity
	if rarity == "mythic" {
		rarity = "mythic rare"
	}

	commanderText := ""
	if isCommander {
		commanderText = "commander"
	}

	searchText := strings.ToLower(fmt.Sprintf("%s, %s, %s, %s, %v, %s, %s",
		scryfallCard.Name,
		scryfallCard.ManaCost,
		scryfallCard.TypeLine,
		scryfallCard.OracleText,
		scryfallCard.Keywords,
		rarity,
		commanderText,
	))

	colors := scryfallCard.Colors
	if colors == nil {
		colors = []string{}
	}
	if len(colors) == 0 {
		colors = append(colors, "C")
	}

	var cardSynergies map[string]float64
	if synergy, ok := (*synergies)[scryfallCard.Name]; ok {
		cardSynergies = synergy
	} else {
		cardSynergies = map[string]float64{}
	}

	price := 0.0
	if parentCard == nil {
		prices := map[scryfallDB.Currency]float64{}
		for currency, currencyPrice := range scryfallCard.Prices {
			prices[currency] = parseAmount(currencyPrice)
		}
		if prices[scryfallDB.USD] > 0.0 {
			//log.Printf("All prices %s: %v", scryfallCard.Name, prices)
			price = prices[scryfallDB.USD]
		} else if prices[scryfallDB.EUR] > 0.0 {
			price = prices[scryfallDB.EUR] * (*priceMap)[scryfallDB.EUR]
		} else if prices[scryfallDB.USD_FOIL] > 0.0 {
			price = prices[scryfallDB.USD_FOIL]
		} else if prices[scryfallDB.EUR_FOIL] > 0.0 {
			price = prices[scryfallDB.EUR_FOIL] * (*priceMap)[scryfallDB.EUR]
		} else if prices[scryfallDB.TIX] > 0.0 {
			price = prices[scryfallDB.TIX] * (*priceMap)[scryfallDB.TIX]
		} else {
			// fmt.Printf("No price found for %s: %v, \n\n", scryfallCard.Name, scryfallCard)
		}
	}

	card := &db.Card{
		ScryfallID:      scryfallCard.ID,
		Name:            scryfallCard.Name,
		Lang:            scryfallCard.Lang,
		ImageURLs:       imageURLs,
		ManaCost:        scryfallCard.ManaCost,
		Cmc:             scryfallCard.Cmc,
		TypeLine:        scryfallCard.TypeLine,
		CardTypes:       cardTypes,
		OracleText:      scryfallCard.OracleText,
		Colors:          colors,
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
		IsLand:          isLand,
		SearchText:      searchText,
		Synergies:       cardSynergies,
	}

	cardgroup.CalculateCardGroups(card)

	return card
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
