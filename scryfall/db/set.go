package db

import (
	"context"
	"errors"
	"log"

	"github.com/maedu/mtg-cards/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ScryfallSet struct {
	ID string `bson:"_id" json:"id"`

	Object        string `json:"object"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	URI           string `json:"uri"`
	ScryfallUri   string `json:"scryfall_uri"`
	SearchUri     string `json:"search_uri"`
	ReleasedAt    string `json:"released_at"`
	SetType       string `json:"set_type"`
	CardCount     int    `json:"card_count"`
	ParentSetCode string `json:"parent_set_code"`
	Digital       bool   `json:"digital"`
	NonfoilOnly   bool   `json:"nonfoil_only"`
	FoilOnnly     bool   `json:"foil_only"`
	IconSvgURI    string `json:"icon_svg_uri"`
}

// ScryfallSetCollection ...
type ScryfallSetCollection struct {
	*mongo.Client
	*mongo.Collection
	context.Context
	context.CancelFunc
}

// GetScryfallSetCollection ...
func GetScryfallSetCollection() ScryfallSetCollection {
	client, ctx, cancel := db.GetConnection()
	db := client.Database(db.GetDatabaseName())
	collection := db.Collection("scryfallsets")

	return ScryfallSetCollection{
		Client:     client,
		Collection: collection,
		Context:    ctx,
		CancelFunc: cancel,
	}
}

// Disconnect ...
func (collection *ScryfallSetCollection) Disconnect() {
	collection.CancelFunc()
	collection.Client.Disconnect(collection.Context)
}

// GetAllScryfallSets Retrives all scryfallsets from the db
func (collection *ScryfallSetCollection) GetAllScryfallSets() ([]*ScryfallSet, error) {
	var scryfallsets []*ScryfallSet = []*ScryfallSet{}
	ctx := collection.Context

	cursor, err := collection.Collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &scryfallsets)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return scryfallsets, nil
}

// GetScryfallSetByID Retrives a scryfallset by its id from the db
func (collection *ScryfallSetCollection) GetScryfallSetByID(id string) (*ScryfallSet, error) {
	var scryfallset *ScryfallSet
	ctx := collection.Context

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Failed parsing id %v", err)
		return nil, err
	}
	result := collection.Collection.FindOne(ctx, bson.D{bson.E{Key: "_id", Value: objID}})
	if result == nil {
		return nil, errors.New("Could not find a ScryfallSet")
	}
	err = result.Decode(&scryfallset)

	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return scryfallset, nil
}

// GetScryfallSetByIDs Retrives scryfallsets by their ids from the db
func (collection *ScryfallSetCollection) GetScryfallSetByIDs(ids *[]primitive.ObjectID) (*[]ScryfallSet, error) {
	ctx := collection.Context
	var scryfallsets []ScryfallSet = []ScryfallSet{}
	cursor, err := collection.Collection.Find(ctx, bson.D{bson.E{Key: "_id", Value: bson.M{"$in": ids}}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &scryfallsets)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return &scryfallsets, nil

}

// GetScryfallSetByKey Retrives a scryfallset by its key from the db
func (collection *ScryfallSetCollection) GetScryfallSetByKey(key string) (*ScryfallSet, error) {
	var scryfallset *ScryfallSet
	ctx := collection.Context

	result := collection.Collection.FindOne(ctx, bson.D{bson.E{Key: "key", Value: key}})
	if result == nil {
		return nil, nil
	}
	err := result.Decode(&scryfallset)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return scryfallset, nil
}

// Create creating a scryfallset in a mongo
func (collection *ScryfallSetCollection) Create(scryfallset *ScryfallSet) (primitive.ObjectID, error) {
	ctx := collection.Context
	scryfallset.ID = primitive.NewObjectID().Hex()

	result, err := collection.Collection.InsertOne(ctx, scryfallset)
	if err != nil {
		log.Printf("Could not create ScryfallSet: %v", err)
		return primitive.NilObjectID, err
	}
	oid := result.InsertedID.(primitive.ObjectID)
	return oid, nil
}

// CreateMany creating many scryfallsets in a mongo
func (collection *ScryfallSetCollection) CreateMany(scryfallsets []*ScryfallSet) (*[]string, error) {
	ctx := collection.Context

	var ui []interface{}
	for _, t := range scryfallsets {
		if t.ID == "" {
			t.ID = primitive.NewObjectID().Hex()
		}
		ui = append(ui, t)
	}

	result, err := collection.Collection.InsertMany(ctx, ui)
	if err != nil {
		log.Printf("Could not create ScryfallSet: %v", err)
		return nil, err
	}

	var oids = []string{}

	for _, id := range result.InsertedIDs {
		oids = append(oids, id.(string))
	}
	return &oids, nil
}

//Update updating an existing scryfallset in a mongo
func (collection *ScryfallSetCollection) Update(scryfallset *ScryfallSet) (*ScryfallSet, error) {
	ctx := collection.Context
	var updatedScryfallSet *ScryfallSet

	update := bson.M{
		"$set": scryfallset,
	}

	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		Upsert:         &upsert,
		ReturnDocument: &after,
	}

	err := collection.Collection.FindOneAndUpdate(ctx, bson.M{"_id": scryfallset.ID}, update, &opt).Decode(&updatedScryfallSet)
	if err != nil {
		log.Printf("Could not save ScryfallSet: %v", err)
		return nil, err
	}
	return updatedScryfallSet, nil
}

// DeleteScryfallSetByID Deletes an scryfallset by its id from the db
func (collection *ScryfallSetCollection) DeleteScryfallSetByID(id string) error {
	ctx := collection.Context
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Failed parsing id %v", err)
		return err
	}
	result, err := collection.Collection.DeleteOne(ctx, bson.D{bson.E{Key: "_id", Value: objID}})
	if result == nil {
		return errors.New("Could not find a ScryfallSet")
	}

	if err != nil {
		log.Printf("Failed deleting %v", err)
		return err
	}
	return nil
}

// ReplaceAll first delete all and then create many cards in a mongo db
func (collection *ScryfallSetCollection) ReplaceAll(cards []*ScryfallSet) (*[]string, error) {
	ctx := collection.Context

	err := collection.Collection.Drop(ctx)
	if err != nil {
		return nil, err
	}

	var ui []interface{}
	for _, t := range cards {
		if t.ID == "" {
			t.ID = primitive.NewObjectID().Hex()
		}
		ui = append(ui, t)
	}

	result, err := collection.Collection.InsertMany(ctx, ui)
	if err != nil {
		log.Printf("Could not create Card: %v", err)
		return nil, err
	}

	var oids = []string{}

	for _, id := range result.InsertedIDs {
		oids = append(oids, id.(string))
	}
	return &oids, nil
}
