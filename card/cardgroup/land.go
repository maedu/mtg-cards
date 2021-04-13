package cardgroup

import (
	"fmt"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateLand(card *db.Card) {
	if isLandOnCardFace(card) {
		card.CardGroups = append(card.CardGroups, "Land")
		card.SearchText = fmt.Sprintf("%s, land", card.SearchText)
	}
}

func isLandOnCardFace(card *db.Card) bool {
	for _, cardType := range card.CardTypes {
		if cardType == db.Land {
			// CardType = Land is handled in cardtypes.go
			return false
		}
	}

	for _, cardFace := range card.CardFaces {
		for _, cardType := range cardFace.CardTypes {
			if cardType == db.Land {
				return true
			}
		}
	}

	return false
}
