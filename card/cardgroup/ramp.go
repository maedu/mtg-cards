package cardgroup

import (
	"fmt"
	"regexp"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateRamp(card *db.Card) {
	if isNonRampCard(card) {
		return
	}
	if hasRampText(card) || isRampCard(card) {
		card.CardGroups = append(card.CardGroups, "Ramp")
		card.SearchText = fmt.Sprintf("%s, ramp", card.SearchText)
	}
}

var rampManaRegex, _ = regexp.Compile(`(?i)Adds? (\{[CBUGRW]\}|[a-zA-Z ]+mana)`)
var landCardRegex, _ = regexp.Compile("(?i)(Land|Forest|Plains|Mountain|Swamp|Island) card[^.]+put .+?onto the battlefield")
var canPlayAdditionalLandRegex, _ = regexp.Compile(`(?i)You may play \w+ additional lands?`)
var canPutAdditionalLandRegex, _ = regexp.Compile(`(?i)You may put a land card from your hand onto the battlefield`)

var rampCards = []string{
	"Jeweled Lotus",
	"Explorer's Scope",
	"Horizon Stone",
}

var nonRampCards = []string{
	"Path to Exile",
	"Settle the Wreckage",
	"Imprisoned in the Moon",
	"Abundant Growth",
	"Urban Utopia",
	"Assassin's Trophy",
}

func hasRampText(card *db.Card) bool {
	for _, cardType := range card.CardTypes {
		if cardType == db.Land {
			return false
		}
	}

	return rampManaRegex.MatchString(card.OracleText) ||
		landCardRegex.MatchString(card.OracleText) ||
		canPlayAdditionalLandRegex.MatchString(card.OracleText) ||
		canPutAdditionalLandRegex.MatchString(card.OracleText)
}

func isRampCard(card *db.Card) bool {
	for _, rampCard := range rampCards {
		if rampCard == card.Name {
			return true
		}
	}
	return false
}

func isNonRampCard(card *db.Card) bool {
	for _, nonRampCard := range nonRampCards {
		if nonRampCard == card.Name {
			return true
		}
	}
	return false
}
