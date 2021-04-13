package cardgroup

import (
	"github.com/maedu/mtg-cards/card/db"
)

func calculateCardTYpes(card *db.Card) {
	for _, cardType := range card.CardTypes {
		card.CardGroups = append(card.CardGroups, string(cardType))
	}
}
