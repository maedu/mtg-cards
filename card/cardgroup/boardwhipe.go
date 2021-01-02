package cardgroup

import (
	"fmt"
	"regexp"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateBoardWhipes(card *db.Card) {
	if isBoardWhipe(card) {
		card.CardGroups = append(card.CardGroups, "Board Whipe")
		card.SearchText = fmt.Sprintf("%s, board whipe", card.SearchText)
	}
}

var boardWhipeDestroyAllRegex, _ = regexp.Compile("Destroy all")

func isBoardWhipe(card *db.Card) bool {
	if card.CardType != db.Instant && card.CardType != db.Sorcery {
		return false
	}

	return boardWhipeDestroyAllRegex.MatchString(card.OracleText)
}
