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
	log.Println("Transform cards")
	for _, scryfallCard := range loadedScryfallCards {
		card := transformCard(scryfallCard, &synergies)
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

func transformCard(scryfallCard *scryfallDB.ScryfallCard, synergies *map[string]map[string]float64) *db.Card {

	switch scryfallCard.Layout {
	case "art_series", "token", "emblem":
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
		cardFace := transformCard(&scryfallCardFace, synergies)
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

	cardTypesToCheck := []db.CardType{
		db.Creature,
		db.Artifact,
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

	card := &db.Card{
		ScryfallID:      scryfallCard.ID,
		Name:            scryfallCard.Name,
		Lang:            scryfallCard.Lang,
		ImageURLs:       imageURLs,
		ManaCost:        scryfallCard.ManaCost,
		Cmc:             scryfallCard.Cmc,
		TypeLine:        scryfallCard.TypeLine,
		CardType:        cardType,
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
