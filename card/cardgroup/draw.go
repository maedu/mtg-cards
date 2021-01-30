package cardgroup

import (
	"fmt"
	"regexp"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateDraw(card *db.Card) {
	if isDraw(card) {
		card.CardGroups = append(card.CardGroups, "Draw")
		card.SearchText = fmt.Sprintf("%s, draw", card.SearchText)
	}
}

var drawRegex, _ = regexp.Compile("(?i)Draw( [a-zA-Z]+)? card")

func isDraw(card *db.Card) bool {
	return drawRegex.MatchString(card.OracleText)
}