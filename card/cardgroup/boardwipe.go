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
var boardWipeExileAllRegex, _ = regexp.Compile("Exile all")

var boardWipeCards = []string{
	"Cyclonic Rift",
	"Duneblast",
	"Toxic Deluge",
	"Time Wipe",
	"Ravnica at War",
	"Nevinyrral's Disk",
	"Star of Extinction",
}

var nonBoardWipeCards = []string{
	"Time Stop",
	"Discontinuity",
	"Synthetic Destiny",
	"Acid Rain",
	"Day's Undoing",
	"Paradigm Shift",
	"Mass Polymorph",
	"Summary Dismissal",
}

func isBoardWipe(card *db.Card) bool {
	for _, nonBoardWipeCard := range nonBoardWipeCards {
		if nonBoardWipeCard == card.Name {
			return false
		}
	}
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

	if boardWipeExileAllRegex.MatchString(card.OracleText) {
		return true
	}

	return false
}
