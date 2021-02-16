package cardgroup

import (
	"fmt"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateLand(card *db.Card) {
	if isLand(card) {
		card.CardGroups = append(card.CardGroups, "Land")
		card.SearchText = fmt.Sprintf("%s, land", card.SearchText)
	}
}

func isLand(card *db.Card) bool {
	if card.CardType == "Land" {
		return true
	}

	for _, cardFace := range card.CardFaces {
		if cardFace.CardType == "Land" {
			return true
		}
	}

	return false
}
