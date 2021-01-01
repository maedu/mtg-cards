package cardgroup

import (
	"fmt"
	"regexp"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateRamp(card *db.Card) {
	if isManaArtifact(card) {
		card.CardGroups = append(card.CardGroups, "Ramp")
		card.SearchText = fmt.Sprintf("%s, ramp", card.SearchText)
	}
}

var rampManaRegex, _ = regexp.Compile("\\{T\\}: Add (\\{[CBUGRW]\\}|[a-zA-Z ]+mana)")

func isManaArtifact(card *db.Card) bool {
	if card.CardType != db.Artifact {
		return false
	}

	return rampManaRegex.MatchString(card.OracleText)
}
