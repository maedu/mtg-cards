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

var boardWipeDestroyAllRegex, _ = regexp.Compile("Destroy all")

var boardWipeCards = []string{
	"Cyclonic Rift",
	"Duneblast",
	"Toxic Deluge",
	"Time Wipe",
	"Ravnica at War",
}

func isBoardWipe(card *db.Card) bool {
	for _, boardWipeCard := range boardWipeCards {
		if boardWipeCard == card.Name {
			return true
		}
	}

	instantSorceryFound := false
	for _, cardType := range card.CardTypes {
		if cardType == db.Instant || cardType == db.Sorcery {
			instantSorceryFound = true
			break
		}
	}

	if !instantSorceryFound {
		return false
	}

	if boardWipeDestroyAllRegex.MatchString(card.OracleText) {
		return true
	}

	return false
}
