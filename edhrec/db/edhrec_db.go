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

type EdhrecSynergy struct {
	ID              string  `bson:"_id" json:"id"`
	MainCard        string  `bson:"main_card" json:"mainCard"`
	CardWithSynergy string  `bson:"card_with_synergy" json:"cardWithSynergy"`
	Synergy         float64 `bson:"synergy" json:"synergty"`
}

// EdhrecSynergyCollection ...
type EdhrecSynergyCollection struct {
	*mongo.Client
	*mongo.Collection
	context.Context
	context.CancelFunc
}

// GetEdhrecSynergyCollection ...
func GetEdhrecSynergyCollection() (EdhrecSynergyCollection, error) {
	client, ctx, cancel := db.GetConnection()
	db := client.Database(db.GetDatabaseName())
	collection := db.Collection("edhrec_synergies")

	model := mongo.IndexModel{
		Keys: bson.M{
			"main_card": 1,
		}, Options: nil,
	}
	_, err := collection.Indexes().CreateOne(ctx, model)

	return EdhrecSynergyCollection{
		Client:     client,
		Collection: collection,
		Context:    ctx,
		CancelFunc: cancel,
	}, err
}

// Disconnect ...
func (collection *EdhrecSynergyCollection) Disconnect() {
	collection.CancelFunc()
	collection.Client.Disconnect(collection.Context)
}

// GetAllEdhrecSynergys Retrives all edhrecsynergys from the db
func (collection *EdhrecSynergyCollection) GetAllEdhrecSynergys() ([]*EdhrecSynergy, error) {
	var edhrecsynergys []*EdhrecSynergy = []*EdhrecSynergy{}
	ctx := collection.Context

	cursor, err := collection.Collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &edhrecsynergys)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return edhrecsynergys, nil
}

// GetEdhrecSynergysByMainCard retrieves edhrecsynergys by their main card from the db
func (collection *EdhrecSynergyCollection) GetEdhrecSynergysByMainCard(mainCard string) ([]EdhrecSynergy, error) {
	ctx := collection.Context
	var edhrecsynergys []EdhrecSynergy = []EdhrecSynergy{}
	cursor, err := collection.Collection.Find(ctx, bson.D{bson.E{Key: "main_card", Value: mainCard}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &edhrecsynergys)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return edhrecsynergys, nil

}

// ReplaceAllOfMainCard first delete all for main card and then create many cards in a mongo db
func (collection *EdhrecSynergyCollection) ReplaceAllOfMainCard(mainCard string, edhrecSynergys []EdhrecSynergy) error {
	ctx := collection.Context

	filter := bson.M{"main_card": bson.M{"$eq": mainCard}}
	result, err := collection.Collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	if result == nil {
		return errors.New("Could not delete a EdhrecSynergys")
	}

	var ui []interface{}
	for _, t := range edhrecSynergys {
		if t.ID == "" {
			t.ID = primitive.NewObjectID().Hex()
		}
		ui = append(ui, t)
	}

	_, err = collection.Collection.InsertMany(ctx, ui)
	return err
}
