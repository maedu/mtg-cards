package db

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/maedu/mtg-cards/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserCard struct {
	ID       string `bson:"_id" json:"-"`
	UserID   string `bson:"user_id" json:"-"`
	Card     string `bson:"card" json:"card"`
	Quantity int64  `bson:"quantity" json:"quantity"`
	Source   string `bson:"source" json:"source"`
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

// UpsertForUserAndSource inserts or updates the user card in a mongo db
func (collection *UserCardCollection) UpsertForUserAndSource(userID string, source string, userCard *UserCard) error {
	ctx := collection.Context

	if userCard.ID == "" {
		userCard.ID = primitive.NewObjectID().Hex()
	}

	filter := bson.M{"$and": []bson.M{
		{"user_id": bson.M{"$eq": userID}},
		{"source": bson.M{"$eq": source}},
		{"card": bson.M{"$eq": userCard.Card}},
	}}

	upsert := true
	options := &options.ReplaceOptions{
		Upsert: &upsert,
	}

	result, err := collection.Collection.ReplaceOne(ctx, filter, userCard, options)

	if err != nil {
		return fmt.Errorf("replaceOne failed: %w", err)
	}
	if result == nil {
		return errors.New("could not delete a UserCards")
	}
	return nil
}

// ReplaceAllOfUserAndSource first delete all for userId & source and then create many cards in a mongo db
func (collection *UserCardCollection) ReplaceAllOfUserAndSource(userID string, source string, userCards []*UserCard) error {
	ctx := collection.Context

	filter := bson.M{"$and": []bson.M{
		{"user_id": bson.M{"$eq": userID}},
		{"source": bson.M{"$eq": source}},
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
