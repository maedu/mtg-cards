package upload

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/maedu/mtg-cards/user/db"
)

type mtgGoldFish struct{}

/*
Example:
Card,Set ID,Set Name,Quantity,Foil
"Forest",DOM,"",1,""
"Island",DOM,"",1,""
"Mountain",DOM,"",1,""
"Plains",DOM,"",1,""
"Swamp",DOM,"",1,""
*/
var mtgGoldFishFields = fields{"Card", "Set ID", "Set Name", "Quantity", "Foil"}

func (u mtgGoldFish) name() string {
	return "MTGGoldfish"
}

func (u mtgGoldFish) parse(reader *csv.Reader, request ParseRequest) (bool, []*db.UserCard, error) {
	headerParsed := false
	// Iterate through the records
	rowCount := 0
	cardMap := map[string]*db.UserCard{}
	for {
		// Read each record from csv
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, nil, fmt.Errorf("reading record failed: %w", err)
		}

		if !headerParsed {
			if record[0] == "Card" { // Row with header fields
				if rowMatchesFields(record, mtgGoldFishFields) {
					headerParsed = true
				} else {
					return false, nil, nil
				}
			}
		} else if allFieldsEmpty(record) {
			// Empty record indicates the end
			break
		} else {

			name := mtgGoldFishFields.value("Card", record)
			quantityS := mtgGoldFishFields.value("Quantity", record)
			quantity, err := parseQuantity(quantityS)
			setID := mtgGoldFishFields.value("Set ID", record)
			if err != nil {
				return false, nil, fmt.Errorf("parsing quantity failed. '%s': %w", quantityS, err)
			}

			if cardMap[name] == nil {
				cardMap[name] = &db.UserCard{
					Name:   name,
					Sets:   []db.Set{},
					UserID: request.UserID,
				}
			}

			cardMap[name].Sets = append(cardMap[name].Sets, db.Set{
				ID:       setID,
				Quantity: quantity,
				Source:   request.Source,
			})

		}

		rowCount++
	}

	if !headerParsed {
		return false, nil, nil
	}

	userCards := []*db.UserCard{}
	for _, card := range cardMap {
		userCards = append(userCards, card)
	}

	return true, userCards, nil
}
