package cardgroup

import "github.com/maedu/mtg-cards/card/db"

// CalculateCardGroups calculates the card groups and updates the card
func CalculateCardGroups(card *db.Card) {
	calculateRamp(card)
	calculateBoardWhipes(card)
	calculateDraw(card)

}
