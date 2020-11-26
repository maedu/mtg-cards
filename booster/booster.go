package booster

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"bitbucket.org/spinnerweb/cards/card/db"
)

// Booster containing cards for an edition
type Booster struct {
	Set   string     `json:"set"`
	Cards []*db.Card `json:"cards"`
}

// GenerateBooster generates and returns a booster pack
func GenerateBooster(set string) (Booster, error) {

	collection, err := db.GetCardCollection()
	if err != nil {
		return Booster{}, err
	}
	defer collection.Disconnect()

	cards, err := collection.GetCardBySetName(set)
	if err != nil {
		return Booster{}, err
	}

	if len(cards) == 0 {
		return Booster{}, fmt.Errorf("No cards found for set %s", set)
	}

	if set == "Commander Legends" {
		return generateCommanderLegendsBooster(cards), nil
	}
	return generateNormalBooster(cards, set), nil
}

func generateCommanderLegendsBooster(cards []*db.Card) Booster {
	// No of cards: 20
	// 1 non-legendary rare/mythic rare
	// 1 foil card (how do we replace that?)
	// 2 legendary creatures
	// Every 6th pack: The Prismatic Piper, which replaces a common

	boosterCards := []*db.Card{}
	boosterCards = append(boosterCards, nonLegendaryRareMythic(cards))
	boosterCards = append(boosterCards, legendaryCreature(cards, 2)...)
	boosterCards = append(boosterCards, uncommonCards(cards, 3)...)
	boosterCards = append(boosterCards, commonCards(cards, 12)...)
	boosterCards = append(boosterCards, prismaticPiperOrCommon(cards))
	boosterCards = append(boosterCards, foil(cards))

	return Booster{
		Cards: boosterCards,
		Set:   "Commander Legends",
	}
}

func generateNormalBooster(cards []*db.Card, set string) Booster {

	boosterCards := []*db.Card{}
	boosterCards = append(boosterCards, rareMythic(cards))
	boosterCards = append(boosterCards, uncommonCards(cards, 3)...)
	boosterCards = append(boosterCards, commonCards(cards, 11)...)

	return Booster{
		Cards: boosterCards,
		Set:   set,
	}
}

func nonLegendaryRareMythic(cards []*db.Card) *db.Card {
	rare := filterByRarity(cards, "rare")
	mythic := filterByRarity(cards, "mythic")
	rareMythic := append(rare, mythic...)
	nonLegendary := filterByNotLegendaryCreature(rareMythic)
	return randomCard(nonLegendary)
}

func rareMythic(cards []*db.Card) *db.Card {
	rare := filterByRarity(cards, "rare")
	mythic := filterByRarity(cards, "mythic")
	rareMythic := append(rare, mythic...)
	return randomCard(rareMythic)
}

func foil(cards []*db.Card) *db.Card {
	uncommon := filterByRarity(cards, "uncommon")
	rare := filterByRarity(cards, "rare")
	mythic := filterByRarity(cards, "mythic")
	uncommonRareMythic := append(uncommon, rare...)
	uncommonRareMythic = append(uncommonRareMythic, mythic...)
	return randomCard(uncommonRareMythic)
}

func legendaryCreature(cards []*db.Card, count int) []*db.Card {
	legendaryCreatures := filterByLegendaryCreature(cards)
	return randomCards(legendaryCreatures, count)
}

func prismaticPiperOrCommon(cards []*db.Card) *db.Card {
	seed := rand.NewSource(time.Now().UnixNano())
	randomizer := rand.New(seed)
	dice := randomizer.Intn(6)
	if dice == 1 {
		return findByName(cards, "The Prismatic Piper")
	}

	commonCards := commonCards(cards, 1)
	return commonCards[0]
}

func commonCards(cards []*db.Card, count int) []*db.Card {
	commonCards := filterByRarity(cards, "common")

	nonBasicLands := []*db.Card{}
	for _, card := range commonCards {
		if !strings.Contains(card.TypeLine, "Basic Land") && !strings.Contains(card.TypeLine, "Basic Snow Land") {
			nonBasicLands = append(nonBasicLands, card)
		}
	}

	return randomCards(nonBasicLands, count)
}

func uncommonCards(cards []*db.Card, count int) []*db.Card {
	uncommonCards := filterByRarity(cards, "uncommon")
	return randomCards(uncommonCards, count)
}

func randomCard(cards []*db.Card) *db.Card {
	seed := rand.NewSource(time.Now().UnixNano())
	randomizer := rand.New(seed)
	randomIndex := randomizer.Intn(len(cards))
	return cards[randomIndex]
}

func randomCards(cards []*db.Card, count int) []*db.Card {
	seed := rand.NewSource(time.Now().UnixNano())
	randomizer := rand.New(seed)

	result := []*db.Card{}
	for i := 0; i < count; i++ {
		randomIndex := randomizer.Intn(len(cards))
		result = append(result, cards[randomIndex])
	}
	return result
}
func filterByRarity(cards []*db.Card, rarity string) []*db.Card {
	result := []*db.Card{}
	for _, card := range cards {
		if card.Rarity == rarity {
			result = append(result, card)
		}
	}

	return result
}

func filterByLegendaryCreature(cards []*db.Card) []*db.Card {
	result := []*db.Card{}
	for _, card := range cards {
		if strings.Contains(card.TypeLine, "Legendary Creature") {
			result = append(result, card)
		}
	}

	return result
}

func filterByNotLegendaryCreature(cards []*db.Card) []*db.Card {
	result := []*db.Card{}
	for _, card := range cards {
		if !strings.Contains(card.TypeLine, "Legendary Creature") {
			result = append(result, card)
		}
	}

	return result
}

func findByName(cards []*db.Card, name string) *db.Card {
	for _, card := range cards {
		if card.Name == name {
			return card
		}
	}

	return nil
}
