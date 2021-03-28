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

var enablerCards = []string{
	"Brave the Elements",
	"Flawless Maneuver",
	"Teferi's Protection",
	"Unbreakable Formation",
	"Lightning Greaves",
	"Swiftfoot Boots",
}

func isEnabler(card *db.Card) bool {
	for _, enablerCard := range enablerCards {
		if enablerCard == card.Name {
			return true
		}
	}
	return false
}
