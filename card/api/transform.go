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
		fmt.Printf("Error updating cards: %v\n", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	err = TransformCards()
	if err != nil {
		fmt.Printf("Error transforming cards: %v\n", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, nil)

}

func handleTransformCards(c *gin.Context) {

	err := TransformCards()
	if err != nil {
		fmt.Printf("Error transforming cards: %v2\n", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, nil)

}

func TransformCards() error {

	log.Println("Get scryfallCollection")
	scryfallCollection := scryfallDB.GetScryfallCardCollection()
	defer scryfallCollection.Disconnect()

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

	collection, err := db.GetCardCollection()
	if err != nil {
		return err
	}
	defer collection.Disconnect()

	collection.DeleteAll()

	priceMap := getPriceMap()

	var page int64 = 0
	var limit int64 = 1000
	first := true

	for first || page > 0 {

		first = false
		loadedScryfallCardsPaginated, err := scryfallCollection.GetScryfallCardsPaginated(limit, page)
		if err != nil {
			return fmt.Errorf("GetScryfallCardsPaginated failed, for page %d: %w", page, err)
		}
		fmt.Printf("Loaded cards, page: %v\n", loadedScryfallCardsPaginated.Pagination.Page)
		cards := []*db.Card{}

		for _, scryfallCard := range loadedScryfallCardsPaginated.Cards {
			card := transformCard(scryfallCard, &synergies, &priceMap, nil)
			if card != nil {
				cards = append(cards, card)
			}
		}
		err = collection.CreateMany(cards)
		if err != nil {
			return fmt.Errorf("CreateMany failed for page %d: %w", page, err)
		}
		page = loadedScryfallCardsPaginated.Pagination.Next
	}
	fmt.Println("Transformation done")
	return nil
}

func getPriceMap() map[scryfallDB.Currency]float64 {
	priceMap := map[scryfallDB.Currency]float64{}
	priceMap[scryfallDB.USD] = 1.0
	priceMap[scryfallDB.EUR] = 1.1812452253628725
	priceMap[scryfallDB.TIX] = 21.747538677918424
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
	if !legalInCommander && parentCard == nil {
		return nil
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
