package cardgroup

import (
	"fmt"

	"github.com/maedu/mtg-cards/card/db"
)

func calculateStandalone(card *db.Card) {
	if isStandalone(card) {
		card.CardGroups = append(card.CardGroups, "Standalone")
		card.SearchText = fmt.Sprintf("%s, standalone", card.SearchText)
	}
}

var StandaloneCards = []string{
	"Authority of the Consuls",
	"Land Tax",
	"Ashes of the Abhorrent",
	"Dawn of Hope",
	"Smothering Tithe",
	"Cosmos Elixir",
	"The Immortal Sun",
}

func isStandalone(card *db.Card) bool {
	for _, StandaloneCard := range StandaloneCards {
		if StandaloneCard == card.Name {
			return true
		}
	}

	return false
}
