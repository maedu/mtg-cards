package cardgroup

import (
	"fmt"
	"regexp"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateRamp(card *db.Card) {
	if hasRampText(card) || isRampCard(card) {
		card.CardGroups = append(card.CardGroups, "Ramp")
		card.SearchText = fmt.Sprintf("%s, ramp", card.SearchText)
	}
}

var rampManaRegex, _ = regexp.Compile("(?i)Adds? (\\{[CBUGRW]\\}|[a-zA-Z ]+mana)")

var rampCards = []string{
	"Jeweled Lotus",
	"Explorer's Scope",
	"Horizon Stone",
}

func hasRampText(card *db.Card) bool {
	if card.CardType == db.Land {
		return false
	}

	return rampManaRegex.MatchString(card.OracleText)
}

func isRampCard(card *db.Card) bool {
	for _, rampCard := range rampCards {
		if rampCard == card.Name {
			return true
		}
	}
	return false
}
