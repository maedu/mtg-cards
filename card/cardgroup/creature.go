package cardgroup

import (
	"fmt"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateCreature(card *db.Card) {
	if isCreature(card) {
		card.CardGroups = append(card.CardGroups, "Creature")
		card.SearchText = fmt.Sprintf("%s, creature", card.SearchText)
	}
}

func isCreature(card *db.Card) bool {
	return card.CardType == db.Creature
}
