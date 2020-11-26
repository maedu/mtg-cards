package game

import (
	"bitbucket.org/spinnerweb/cards/booster"
	"bitbucket.org/spinnerweb/cards/card/db"
)

// State defines the state of the game
type State string

const (
	// WaitingForPlayers indicates that the game is still open and players can join
	WaitingForPlayers State = "Waiting for players"
	// Started indicates that the game has started, no more players can join
	Started State = "Started"
)

// Game stores the status of a game, such as players, booster pool etc
type Game struct {
	Key         string
	Set         string
	Players     []*Player
	BoosterPool []*booster.Booster
	State       State
}

// Player participates in the game
type Player struct {
	Name     string
	Boosters []*booster.Booster
	Cards    []*db.Card
}

// InitGame initializes a new game
func InitGame(key string, set string) Game {
	return Game{
		Key:         key,
		Set:         set,
		Players:     []*Player{},
		BoosterPool: []*booster.Booster{},
		State:       WaitingForPlayers,
	}
}

// AddPlayer adds a player to the game
func (game Game) AddPlayer(name string) error {
	player := &Player{
		Name:     name,
		Boosters: []*booster.Booster{},
		Cards:    []*db.Card{},
	}

	game.Players = append(game.Players, player)
	boosters, err := getBoostersForPlayer(game.Set)
	if err != nil {
		return err
	}
	game.BoosterPool = append(game.BoosterPool, boosters...)
	return nil
}

func getBoostersForPlayer(set string) ([]*booster.Booster, error) {
	boosters := []*booster.Booster{}
	for i := 0; i < 3; i++ {
		booster, err := booster.GenerateBooster(set)
		if err != nil {
			return nil, err
		}

		boosters = append(boosters, &booster)
	}
	return boosters, nil
}

// Start the game
func (game Game) Start() {
	for _, player := range game.Players {
		player.Boosters = append(player.Boosters, game.BoosterPool[0])
		game.BoosterPool = game.BoosterPool[1:]
	}
	game.State = Started
}
