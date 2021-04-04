package cardgroup

import "github.com/maedu/mtg-cards/card/db"

// CalculateCardGroups calculates the card groups and updates the card
func CalculateCardGroups(card *db.Card) {
	card.CardGroups = []string{}
	calculateRamp(card)
	calculateDraw(card)
	calculateBoardWipes(card)
	calculateRemoval(card)
	calculateLand(card)
	calculateCreature(card)
	calculateStandalone(card)
	calculateEnhancers(card)
	calculateEnablers(card)
	calculateCommander(card)

}
