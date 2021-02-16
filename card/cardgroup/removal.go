package cardgroup

import (
	"fmt"
	"regexp"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateRemoval(card *db.Card) {
	if isRemoval(card) {
		card.CardGroups = append(card.CardGroups, "Removal")
		card.SearchText = fmt.Sprintf("%s, removal", card.SearchText)
	}
}

var removalDestroyTargetRegex, _ = regexp.Compile("(?i)Destroy target")

var removalExile, _ = regexp.Compile("(?i)exile( another)? target")
var removalExileGraveyard, _ = regexp.Compile("(?i)exile( another)? target([^.])+graveyard")

var removalCards = []string{
	"Duneblast",
}

func isRemoval(card *db.Card) bool {
	if removalDestroyTargetRegex.MatchString(card.OracleText) {
		return true
	}
	if removalExile.MatchString(card.OracleText) {
		return !removalExileGraveyard.MatchString(card.OracleText)
	}
	for _, removalCard := range removalCards {
		if removalCard == card.Name {
			return true
		}
	}
	return false
}
