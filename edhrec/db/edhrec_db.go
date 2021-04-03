package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	pagination "github.com/maedu/mongo-go-pagination"
	"github.com/maedu/mtg-cards/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Set string
type Rarity string

type PaginatedResult struct {
	Edhrecs    []*Edhrec                 `json:"edhrecs"`
	Pagination pagination.PaginationData `json:"pagination"`
}

type EdhrecType string

const (
	Artifact     EdhrecType = "Artifact"
	Conspiracy   EdhrecType = "Conspiracy"
	Creature     EdhrecType = "Creature"
	Enchantment  EdhrecType = "Enchantment"
	Instant      EdhrecType = "Instant"
	Land         EdhrecType = "Land"
	Phenomenon   EdhrecType = "Phenomenon"
	Plane        EdhrecType = "Plane"
	Planeswalker EdhrecType = "Planeswalker"
	Scheme       EdhrecType = "Scheme"
	Sorcery      EdhrecType = "Sorcery"
	Tribal       EdhrecType = "Tribal"
	Vanguard     EdhrecType = "Vanguard"
)

type Edhrec struct {
	ID              primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	CreatedAt       time.Time          `bson:"created_at" json:"-"`
	RelatedCardName string             `bson:"related_card_name" json:"relatedCardName"`
	Name            string             `bson:"name" json:"name"`
	Synergy         float64            `bson:"-" json:"synergy"`
}

// EdhrecCollection ...
type EdhrecCollection struct {
	*mongo.Client
	*mongo.Collection
	context.Context
	context.CancelFunc
}

// GetEdhrecCollection ...
func GetEdhrecCollection() (EdhrecCollection, error) {
	client, ctx, cancel := db.GetConnection()
	db := client.Database(db.GetDatabaseName())
	collection := db.Collection("edhrecs")
	modelOpts := options.Index()
	modelOpts.SetWeights(bson.M{
		"expire_at":     5,
		"oracle_text":   4,
		"type_line":     2,
		"edhrec_groups": 2,
		"set_name":      1,
	})
	index := mongo.IndexModel{
		Keys:        []string{"expire_at"},
		ExpireAfter: 3600,
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	indexName, err := collection.Indexes().CreateOne(ctx, index, opts)
	if err != nil {
		return EdhrecCollection{}, err
	}
	fmt.Printf("Index created: %s\n", indexName)

	return EdhrecCollection{
		Client:     client,
		Collection: collection,
		Context:    ctx,
		CancelFunc: cancel,
	}, nil
}

// Disconnect ...
func (collection *EdhrecCollection) Disconnect() {
	collection.CancelFunc()
	collection.Client.Disconnect(collection.Context)
}

// GetAllEdhrecs Retrives all edhrecs from the db
func (collection *EdhrecCollection) GetAllEdhrecs() ([]*Edhrec, error) {
	var edhrecs []*Edhrec = []*Edhrec{}
	ctx := collection.Context

	cursor, err := collection.Collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &edhrecs)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return edhrecs, nil
}

// GetEdhrecByID Retrives a edhrec by its id from the db
func (collection *EdhrecCollection) GetEdhrecByID(id string) (*Edhrec, error) {
	var edhrec *Edhrec
	ctx := collection.Context

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Failed parsing id %v", err)
		return nil, err
	}
	result := collection.Collection.FindOne(ctx, bson.D{bson.E{Key: "_id", Value: objID}})
	if result == nil {
		return nil, errors.New("Could not find a Edhrec")
	}
	err = result.Decode(&edhrec)

	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return edhrec, nil
}

// GetEdhrecByIDs Retrives edhrecs by their ids from the db
func (collection *EdhrecCollection) GetEdhrecByIDs(ids *[]primitive.ObjectID) (*[]Edhrec, error) {
	ctx := collection.Context
	var edhrecs []Edhrec = []Edhrec{}
	cursor, err := collection.Collection.Find(ctx, bson.D{bson.E{Key: "_id", Value: bson.M{"$in": ids}}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &edhrecs)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return &edhrecs, nil

}

// GetEdhrecByName retrives a edhrec by its key from the db
func (collection *EdhrecCollection) GetEdhrecByName(name string) (*Edhrec, error) {
	var edhrec *Edhrec
	ctx := collection.Context

	filter := bson.M{"$or": bson.A{
		bson.M{"name": name},
		bson.M{"edhrec_faces.name": name},
	}}

	result := collection.Collection.FindOne(ctx, filter)
	if result == nil {
		return nil, nil
	}
	err := result.Decode(&edhrec)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return edhrec, nil
}

// GetEdhrecsByNames retrieves edhrecs by their names from the db
func (collection *EdhrecCollection) GetEdhrecsByNames(names []string) ([]*Edhrec, error) {
	log.Printf("find edhrecs by names: %d", len(names))
	var edhrecs []*Edhrec = []*Edhrec{}
	ctx := collection.Context

	inFilter := bson.M{"$in": names}

	filter := bson.M{"$or": bson.A{
		bson.M{"name": inFilter},
		bson.M{"edhrec_faces.name": inFilter},
	}}

	cursor, err := collection.Collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("Find failed: %w", err)
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &edhrecs)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return edhrecs, nil
}

// GetEdhrecsBySetName retrieves a edhrec by its set name from the db
func (collection *EdhrecCollection) GetEdhrecsBySetName(setName string) ([]*Edhrec, error) {
	var edhrecs []*Edhrec = []*Edhrec{}
	ctx := collection.Context

	cursor, err := collection.Collection.Find(ctx, bson.D{bson.E{Key: "set_name", Value: setName}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &edhrecs)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return edhrecs, nil
}

// Create creating a edhrec in a mongo
func (collection *EdhrecCollection) Create(edhrec *Edhrec) (primitive.ObjectID, error) {
	ctx := collection.Context
	edhrec.ID = primitive.NewObjectID()

	result, err := collection.Collection.InsertOne(ctx, edhrec)
	if err != nil {
		log.Printf("Could not create Edhrec: %v", err)
		return primitive.NilObjectID, err
	}
	oid := result.InsertedID.(primitive.ObjectID)
	return oid, nil
}

// CreateMany creating many edhrecs in a mongo
func (collection *EdhrecCollection) CreateMany(edhrecs []*Edhrec) (*[]string, error) {
	ctx := collection.Context

	var ui []interface{}
	for _, t := range edhrecs {
		t.ID = primitive.NewObjectID()
		ui = append(ui, t)
	}

	result, err := collection.Collection.InsertMany(ctx, ui)
	if err != nil {
		log.Printf("Could not create Edhrec: %v", err)
		return nil, err
	}

	var oids = []string{}

	for _, id := range result.InsertedIDs {
		oids = append(oids, id.(string))
	}
	return &oids, nil
}

//Update updating an existing edhrec in a mongo
func (collection *EdhrecCollection) Update(edhrec *Edhrec) (*Edhrec, error) {
	ctx := collection.Context
	var updatedEdhrec *Edhrec

	update := bson.M{
		"$set": edhrec,
	}

	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		Upsert:         &upsert,
		ReturnDocument: &after,
	}

	err := collection.Collection.FindOneAndUpdate(ctx, bson.M{"_id": edhrec.ID}, update, &opt).Decode(&updatedEdhrec)
	if err != nil {
		log.Printf("Could not save Edhrec: %v", err)
		return nil, err
	}
	return updatedEdhrec, nil
}

// DeleteEdhrecByID Deletes an edhrec by its id from the db
func (collection *EdhrecCollection) DeleteEdhrecByID(id string) error {
	ctx := collection.Context
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Failed parsing id %v", err)
		return err
	}
	result, err := collection.Collection.DeleteOne(ctx, bson.D{bson.E{Key: "_id", Value: objID}})
	if result == nil {
		return errors.New("Could not find a Edhrec")
	}

	if err != nil {
		log.Printf("Failed deleting %v", err)
		return err
	}
	return nil
}

// ReplaceAll first delete all and then create many edhrecs in a mongo db
func (collection *EdhrecCollection) ReplaceAll(edhrecs []*Edhrec) (*[]primitive.ObjectID, error) {
	ctx := collection.Context

	err := collection.Collection.Drop(ctx)
	if err != nil {
		return nil, err
	}

	var ui []interface{}
	for _, t := range edhrecs {
		t.ID = primitive.NewObjectID()
		ui = append(ui, t)
	}

	result, err := collection.Collection.InsertMany(ctx, ui)
	if err != nil {
		log.Printf("Could not create Edhrec: %v", err)
		return nil, err
	}

	var oids = []primitive.ObjectID{}

	for _, id := range result.InsertedIDs {
		oids = append(oids, id.(primitive.ObjectID))
	}
	return &oids, nil
}
