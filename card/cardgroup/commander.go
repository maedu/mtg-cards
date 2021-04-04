package cardgroup

import (
	"fmt"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateCommander(card *db.Card) {
	if card.IsCommander {
		card.CardGroups = append(card.CardGroups, "Commander")
		card.SearchText = fmt.Sprintf("%s, commander", card.SearchText)
	}
}
