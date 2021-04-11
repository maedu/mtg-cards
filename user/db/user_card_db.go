package db

import (
	"context"
	"errors"
	"log"

	"github.com/maedu/mtg-cards/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserCard struct {
	ID           string `bson:"_id" json:"-"`
	UserID       string `bson:"user_id" json:"-"`
	Card         string `bson:"card" json:"card"`
	SetID        string `bson:"set_id" json:"setId"`
	CardAndSetID string `bson:"card_and_set_id" json:"-"`
	Quantity     int64  `bson:"quantity" json:"quantity"`
	Source       string `bson:"source" json:"source"`
}

// UserCardCollection ...
type UserCardCollection struct {
	*mongo.Client
	*mongo.Collection
	context.Context
	context.CancelFunc
}

// GetUserCardCollection ...
func GetUserCardCollection() (UserCardCollection, error) {
	client, ctx, cancel := db.GetConnection()
	db := client.Database(db.GetDatabaseName())
	collection := db.Collection("user_cards")

	model := mongo.IndexModel{
		Keys: bson.M{
			"user_id": 1,
		}, Options: nil,
	}
	_, err := collection.Indexes().CreateOne(ctx, model)

	return UserCardCollection{
		Client:     client,
		Collection: collection,
		Context:    ctx,
		CancelFunc: cancel,
	}, err
}

// Disconnect ...
func (collection *UserCardCollection) Disconnect() {
	collection.CancelFunc()
	collection.Client.Disconnect(collection.Context)
}

// GetAllUserCards Retrives all usercards from the db
func (collection *UserCardCollection) GetAllUserCards() ([]*UserCard, error) {
	var usercards []*UserCard = []*UserCard{}
	ctx := collection.Context

	cursor, err := collection.Collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &usercards)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return usercards, nil
}

// GetUserCardsByUserID retrieves usercards for the user from the db
func (collection *UserCardCollection) GetUserCardsByUserID(userID string) ([]UserCard, error) {
	ctx := collection.Context
	var usercards []UserCard = []UserCard{}
	cursor, err := collection.Collection.Find(ctx, bson.D{bson.E{Key: "user_id", Value: userID}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &usercards)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return usercards, nil

}

// ReplaceAllOfUserAndSource first delete all for userId & source and then create many cards in a mongo db
func (collection *UserCardCollection) ReplaceAllOfUserAndSource(userID string, source string, userCards []*UserCard) error {
	ctx := collection.Context

	cardAndSetIDs := []string{}
	for _, card := range userCards {
		cardAndSetID := getCardAndSetID(card)
		cardAndSetIDs = append(cardAndSetIDs, cardAndSetID)
		card.CardAndSetID = cardAndSetID
	}

	filter := bson.M{"$and": []bson.M{
		{"user_id": bson.M{"$eq": userID}},
		{"source": bson.M{"$eq": source}},
		{"card_and_set_id": bson.M{"$in": cardAndSetIDs}},
	}}

	result, err := collection.Collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	if result == nil {
		return errors.New("could not delete a UserCards")
	}

	var ui []interface{}
	for _, t := range userCards {
		if t.ID == "" {
			t.ID = primitive.NewObjectID().Hex()
		}
		ui = append(ui, t)
	}

	_, err = collection.Collection.InsertMany(ctx, ui)
	return err
}

func getCardAndSetID(card *UserCard) string {
	return card.Card + " " + card.SetID
}
