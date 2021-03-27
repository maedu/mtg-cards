package cardgroup

import (
	"fmt"
	"regexp"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateBoardWipes(card *db.Card) {
	if isBoardWipe(card) {
		card.CardGroups = append(card.CardGroups, "Board Wipe")
		card.SearchText = fmt.Sprintf("%s, board wipe", card.SearchText)
	}
}

var BoardWipeDestroyAllRegex, _ = regexp.Compile("Destroy all")

var BoardWipeCards = []string{
	"Cyclonic Rift",
	"Duneblast",
	"Toxic Deluge",
	"Time Wipe",
}

func isBoardWipe(card *db.Card) bool {
	if card.CardType != db.Instant && card.CardType != db.Sorcery {
		return false
	}

	if BoardWipeDestroyAllRegex.MatchString(card.OracleText) {
		return true
	}

	for _, BoardWipeCard := range BoardWipeCards {
		if BoardWipeCard == card.Name {
			return true
		}
	}

	return false
}
