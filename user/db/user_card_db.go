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

type Set struct {
	ID       string `bson:"id" json:"id"`
	Quantity int64  `bson:"quantity" json:"quantity"`
	Source   string `bson:"source" json:"source"`
}

type UserCard struct {
	ID     string `bson:"_id" json:"-"`
	UserID string `bson:"user_id" json:"-"`
	Name   string `bson:"name" json:"name"`
	Sets   []Set  `bson:"sets" json:"sets"`
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
func (collection *UserCardCollection) GetUserCardsByUserID(userID string) ([]*UserCard, error) {
	ctx := collection.Context
	var usercards []*UserCard = []*UserCard{}
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
func (collection *UserCardCollection) ReplaceAllOfUserAndSourceNewBad(userID string, source string, userCards []*UserCard) error {
	ctx := collection.Context

	models := []mongo.WriteModel{}
	upsert := true
	for _, card := range userCards {
		filter := bson.M{"$and": []bson.M{
			{"user_id": bson.M{"$eq": userID}},
			{"name": bson.M{"$eq": card.Name}},
		}}

		for _, cardSet := range card.Sets {

			update := bson.M{"$set": bson.M{
				"sets.$[element]": cardSet,
			}}

			cardSetForFilter := []bson.M{
				{"id": cardSet.ID},
				{"source": cardSet.Source},
			}

			arrayFilters := []interface{}{bson.M{"element": cardSetForFilter}}

			model := mongo.UpdateOneModel{
				Filter: filter,
				Upsert: &upsert,
				Update: update,
				ArrayFilters: &options.ArrayFilters{
					Filters: arrayFilters,
				},
			}

			models = append(models, &model)
		}

		break
	}

	/*
			db.collection.bulkWrite( [
		   { updateOne :
		      {
		         "filter": <document>,
		         "update": <document or pipeline>,            // Changed in 4.2
		         "upsert": <boolean>,
		         "collation": <document>,                     // Available starting in 3.4
		         "arrayFilters": [ <filterdocument1>, ... ],  // Available starting in 3.6
		         "hint": <document|string>                    // Available starting in 4.2.1
		      }
		   }
		] )
	*/

	ordered := false
	options := &options.BulkWriteOptions{
		Ordered: &ordered,
	}

	result, err := collection.Collection.BulkWrite(ctx, models, options)
	if err != nil {
		fmt.Printf("Error in writing: %v\n", err)
		return err
	}
	if result == nil {
		return errors.New("could not bulkWrite a UserCards")
	}
	fmt.Printf("BulkWrite Result: %v", result)
	return nil
}

// ReplaceAllOfUser first delete all for userId & source and then create many cards in a mongo db
func (collection *UserCardCollection) ReplaceAllOfUser(userID string, userCards []*UserCard) error {
	ctx := collection.Context

	filter := bson.M{"user_id": bson.M{"$eq": userID}}

	result, err := collection.Collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	if result == nil {
		return errors.New("could not delete UserCards")
	}

	var ui []interface{}
	for _, t := range userCards {
		t.ID = primitive.NewObjectID().Hex()
		ui = append(ui, t)
	}

	_, err = collection.Collection.InsertMany(ctx, ui)
	return err
}
