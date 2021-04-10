package cardgroup

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateDraw(card *db.Card) {
	if isDraw(card) {
		card.CardGroups = append(card.CardGroups, "Draw")
		card.SearchText = fmt.Sprintf("%s, draw", card.SearchText)
	}
}

var drawRegex, _ = regexp.Compile("(?i)Draw( [a-zA-Z]+| that many)? card")

var nonCardDraw = []string{
	"Hullbreacher",
	"Veil of Summer",
	"The Locust God",
	"Chasm Skulker",
}

func isDraw(card *db.Card) bool {
	sanitizedOracleText := removeNonDrawTextFromOracleText(card)
	return drawRegex.MatchString(sanitizedOracleText) && !matchNonCardDraw(card)
}

func removeNonDrawTextFromOracleText(card *db.Card) string {
	oracleText := card.OracleText
	oracleText = strings.ReplaceAll(oracleText, "Discard this card: Draw a card.", "")
	oracleText = strings.ReplaceAll(oracleText, "Draw a card, then discard a card.", "")
	oracleText = strings.ReplaceAll(oracleText, fmt.Sprintf("Sacrifice %s: Draw a card.", card.Name), "")

	if strings.Contains(card.TypeLine, "Sorcery") || strings.Contains(card.TypeLine, "Instant") {
		oracleText = strings.ReplaceAll(oracleText, "Draw a card.", "")
		oracleText = strings.ReplaceAll(oracleText, "draw a card.", "")
		oracleText = strings.ReplaceAll(oracleText, "Draw a card at the beginning", "")
	}

	return oracleText
}

func matchNonCardDraw(card *db.Card) bool {
	for _, nonCardDrawCard := range nonCardDraw {
		if nonCardDrawCard == card.Name {
			return true
		}
	}
	return false
}
