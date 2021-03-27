package cardgroup

import (
	"fmt"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateEnablers(card *db.Card) {
	if isEnabler(card) {
		card.CardGroups = append(card.CardGroups, "Enabler")
		card.SearchText = fmt.Sprintf("%s, enabler", card.SearchText)
	}
}

var EnablerCards = []string{
	"Brave the Elements",
	"Flawless Maneuver",
	"Teferi's Protection",
	"Unbreakable Formation",
	"Lightning Greaves",
	"Swiftfoot Boots",
}

func isEnabler(card *db.Card) bool {
	for _, EnablerCard := range EnablerCards {
		if EnablerCard == card.Name {
			return true
		}
	}
	return false
}
