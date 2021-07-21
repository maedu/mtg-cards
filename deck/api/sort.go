package api

import "github.com/maedu/mtg-cards/card/db"

// ByCommanders implements sort.Interface based on the Age field.
type ByName []*db.Card

func (a ByName) Len() int           { return len(a) }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// ByCommanders implements sort.Interface based on the Age field.
type ByCommanders []Deck

func (a ByCommanders) Len() int { return len(a) }
func (a ByCommanders) Less(i, j int) bool {
	if len(a[i].Commanders) == 0 {
		return false
	}
	if len(a[j].Commanders) == 0 {
		return true
	}
	return a[i].Commanders[0].Name < a[j].Commanders[0].Name
}
func (a ByCommanders) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
