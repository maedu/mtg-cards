package db

import (
	"context"
	"log"

	cardDB "github.com/maedu/mtg-cards/card/db"
	"github.com/maedu/mtg-cards/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Settings struct {
	DeckType         string `bson:"type" json:"type"`
	Lands            int    `bson:"lands" json:"lands"`
	SnowCoveredLands bool   `bson:"snowCoveredLands" json:"snowCoveredLands"`
}

type Deck struct {
	ID          string        `bson:"_id" json:"-"`
	UserID      string        `bson:"user_id" json:"-"`
	URLHash     string        `bson:"urlHash" json:"urlHash"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
	Commanders  []cardDB.Card `bson:"commanders" json:"commanders"`
	Cards       []cardDB.Card `bson:"cards" json:"cards"`
	Library     []cardDB.Card `bson:"library" json:"library"`
	Settings    Settings      `bson:"settings" json:"settings"`
}

// DeckCollection ...
type DeckCollection struct {
	*mongo.Client
	*mongo.Collection
	context.Context
	context.CancelFunc
}

// GetDeckCollection ...
func GetDeckCollection() (DeckCollection, error) {
	client, ctx, cancel := db.GetConnection()
	db := client.Database(db.GetDatabaseName())
	collection := db.Collection("decks")

	model := mongo.IndexModel{
		Keys: bson.M{
			"user_id": 1,
		}, Options: nil,
	}
	_, err := collection.Indexes().CreateOne(ctx, model)

	return DeckCollection{
		Client:     client,
		Collection: collection,
		Context:    ctx,
		CancelFunc: cancel,
	}, err
}

// Disconnect ...
func (collection *DeckCollection) Disconnect() {
	collection.CancelFunc()
	collection.Client.Disconnect(collection.Context)
}

// GetAllDecks Retrives all decks from the db
func (collection *DeckCollection) GetAllDecks() ([]*Deck, error) {
	var decks []*Deck = []*Deck{}
	ctx := collection.Context

	cursor, err := collection.Collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &decks)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return decks, nil
}

// GetDecksByUserID retrieves decks for the user from the db
func (collection *DeckCollection) GetDecksByUserID(userID string) ([]*Deck, error) {
	ctx := collection.Context
	var decks []*Deck = []*Deck{}
	cursor, err := collection.Collection.Find(ctx, bson.D{bson.E{Key: "user_id", Value: userID}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &decks)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return decks, nil

}

// GetDeckByKey Retrives a scryfallcard by its key from the db
func (collection *DeckCollection) GetDeckByURLHash(urlHash string) (*Deck, error) {
	var deck *Deck
	ctx := collection.Context

	result := collection.Collection.FindOne(ctx, bson.D{bson.E{Key: "urlHash", Value: urlHash}})
	if result == nil {
		return nil, nil
	}
	err := result.Decode(&deck)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return deck, nil
}
