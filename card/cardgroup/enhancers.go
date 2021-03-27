package cardgroup

import (
	"fmt"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateEnhancers(card *db.Card) {
	if isEnhancer(card) {
		card.CardGroups = append(card.CardGroups, "Enhancer")
		card.SearchText = fmt.Sprintf("%s, enhancer", card.SearchText)
	}
}

var EnhancerCards = []string{
	"Ajani's Welcome",
	"Crusade",
	"Honor of the Pure",
	"Intangible Virtue",
	"Glorious Anthem",
	"Heliod, Sun-Crowned",
	"Anointed Procession",
	"Cathars' Crusade",
	"Divine Visitation",
	"Skullclamp",
	"Hall of Triumph",
	"Heraldic Banner",
	"Coat of Arms",
	"Nyx Lotus",
	"Well of Lost Dreams",
	"Coat of Arms",
	"Gauntlet of Power",
	"Caged Sun",
}

func isEnhancer(card *db.Card) bool {
	for _, EnhancerCard := range EnhancerCards {
		if EnhancerCard == card.Name {
			return true
		}
	}

	return false
}
