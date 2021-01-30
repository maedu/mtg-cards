package db

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	pagination "github.com/maedu/mongo-go-pagination"
	"github.com/maedu/mtg-cards/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Rarity string

type PaginatedResult struct {
	Sets       []*Set                    `json:"sets"`
	Pagination pagination.PaginationData `json:"pagination"`
}

type SetType string

const (
	Artifact     SetType = "Artifact"
	Conspiracy   SetType = "Conspiracy"
	Creature     SetType = "Creature"
	Enchantment  SetType = "Enchantment"
	Instant      SetType = "Instant"
	Land         SetType = "Land"
	Phenomenon   SetType = "Phenomenon"
	Plane        SetType = "Plane"
	Planeswalker SetType = "Planeswalker"
	Scheme       SetType = "Scheme"
	Sorcery      SetType = "Sorcery"
	Tribal       SetType = "Tribal"
	Vanguard     SetType = "Vanguard"
)

type Set struct {
	ID         primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	ScryfallID string             `bson:"scryfall_id" json:"-"`
	Name       string             `json:"name"`
	Code       string             `json:"code"`
	SetType    string             `json:"setType"`
	ReleasedAt time.Time          `json:"releasedAt"`
	CardCount  int                `json:"cardCount"`
	IconURL    string             `json:"iconUrl"`
}

// SetCollection ...
type SetCollection struct {
	*mongo.Client
	*mongo.Collection
	context.Context
	context.CancelFunc
}

// GetSetCollection ...
func GetSetCollection() (SetCollection, error) {
	client, ctx, cancel := db.GetConnection()
	db := client.Database(db.GetDatabaseName())
	collection := db.Collection("sets")

	return SetCollection{
		Client:     client,
		Collection: collection,
		Context:    ctx,
		CancelFunc: cancel,
	}, nil
}

// Disconnect ...
func (collection *SetCollection) Disconnect() {
	collection.CancelFunc()
	collection.Client.Disconnect(collection.Context)
}

// GetAllSets Retrives all sets from the db
func (collection *SetCollection) GetAllSets() ([]*Set, error) {
	var sets []*Set = []*Set{}
	ctx := collection.Context

	cursor, err := collection.Collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &sets)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return sets, nil
}

// GetSetsPaginated Retrives all sets from the db
func (collection *SetCollection) GetSetsPaginated(limit int64, page int64, filterByFullText string) (PaginatedResult, error) {
	var sets []*Set = []*Set{}
	filter := bson.M{}

	trimmedFilterByFullText := strings.TrimSpace(filterByFullText)
	projection := bson.M{}
	var sort interface{} = bson.D{
		{"name", 1},
	}
	if trimmedFilterByFullText != "" {
		filter = bson.M{
			"$text": bson.M{
				"$search": trimmedFilterByFullText,
			},
		}

		projection = bson.M{
			"score": bson.M{
				"$meta": "textScore",
			},
		}

		sort = bson.M{
			"score": bson.M{
				"$meta": "textScore",
			},
		}
	}
	paginatedData, err := pagination.New(collection.Collection).Limit(limit).Page(page).Filter(filter).Select(projection).Sort(sort).Find()
	if err != nil {
		return PaginatedResult{}, err
	}

	for _, raw := range paginatedData.Data {
		var set *Set
		if marshallErr := bson.Unmarshal(raw, &set); marshallErr != nil {
			log.Printf("Failed marshalling: %v", err)
			return PaginatedResult{}, err
		}
		sets = append(sets, set)

	}
	return PaginatedResult{
		Sets:       sets,
		Pagination: paginatedData.Pagination,
	}, nil
}

// GetSetByID Retrives a set by its id from the db
func (collection *SetCollection) GetSetByID(id string) (*Set, error) {
	var set *Set
	ctx := collection.Context

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Failed parsing id %v", err)
		return nil, err
	}
	result := collection.Collection.FindOne(ctx, bson.D{bson.E{Key: "_id", Value: objID}})
	if result == nil {
		return nil, errors.New("Could not find a Set")
	}
	err = result.Decode(&set)

	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return set, nil
}

// GetSetByIDs Retrives sets by their ids from the db
func (collection *SetCollection) GetSetByIDs(ids *[]primitive.ObjectID) (*[]Set, error) {
	ctx := collection.Context
	var sets []Set = []Set{}
	cursor, err := collection.Collection.Find(ctx, bson.D{bson.E{Key: "_id", Value: bson.M{"$in": ids}}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &sets)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return &sets, nil

}

// GetSetByName retrives a set by its key from the db
func (collection *SetCollection) GetSetByName(name string) (*Set, error) {
	var set *Set
	ctx := collection.Context

	filter := bson.M{"$or": bson.A{
		bson.M{"name": name},
		bson.M{"set_faces.name": name},
	}}

	result := collection.Collection.FindOne(ctx, filter)
	if result == nil {
		return nil, nil
	}
	err := result.Decode(&set)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return set, nil
}

// GetSetsByNames retrieves sets by their names from the db
func (collection *SetCollection) GetSetsByNames(names []string) ([]*Set, error) {
	log.Println("find sets by names")
	var sets []*Set = []*Set{}
	ctx := collection.Context

	inFilter := bson.M{"$in": names}

	filter := bson.M{"$or": bson.A{
		bson.M{"name": inFilter},
		bson.M{"set_faces.name": inFilter},
	}}

	cursor, err := collection.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &sets)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return sets, nil
}

// GetSetsBySetName retrieves a set by its set name from the db
func (collection *SetCollection) GetSetsBySetName(setName string) ([]*Set, error) {
	var sets []*Set = []*Set{}
	ctx := collection.Context

	cursor, err := collection.Collection.Find(ctx, bson.D{bson.E{Key: "set_name", Value: setName}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &sets)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return sets, nil
}

// Create creating a set in a mongo
func (collection *SetCollection) Create(set *Set) (primitive.ObjectID, error) {
	ctx := collection.Context
	set.ID = primitive.NewObjectID()

	result, err := collection.Collection.InsertOne(ctx, set)
	if err != nil {
		log.Printf("Could not create Set: %v", err)
		return primitive.NilObjectID, err
	}
	oid := result.InsertedID.(primitive.ObjectID)
	return oid, nil
}

// CreateMany creating many sets in a mongo
func (collection *SetCollection) CreateMany(sets []*Set) (*[]string, error) {
	ctx := collection.Context

	var ui []interface{}
	for _, t := range sets {
		t.ID = primitive.NewObjectID()
		ui = append(ui, t)
	}

	result, err := collection.Collection.InsertMany(ctx, ui)
	if err != nil {
		log.Printf("Could not create Set: %v", err)
		return nil, err
	}

	var oids = []string{}

	for _, id := range result.InsertedIDs {
		oids = append(oids, id.(string))
	}
	return &oids, nil
}

//Update updating an existing set in a mongo
func (collection *SetCollection) Update(set *Set) (*Set, error) {
	ctx := collection.Context
	var updatedSet *Set

	update := bson.M{
		"$set": set,
	}

	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		Upsert:         &upsert,
		ReturnDocument: &after,
	}

	err := collection.Collection.FindOneAndUpdate(ctx, bson.M{"_id": set.ID}, update, &opt).Decode(&updatedSet)
	if err != nil {
		log.Printf("Could not save Set: %v", err)
		return nil, err
	}
	return updatedSet, nil
}

// DeleteSetByID Deletes an set by its id from the db
func (collection *SetCollection) DeleteSetByID(id string) error {
	ctx := collection.Context
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Failed parsing id %v", err)
		return err
	}
	result, err := collection.Collection.DeleteOne(ctx, bson.D{bson.E{Key: "_id", Value: objID}})
	if result == nil {
		return errors.New("Could not find a Set")
	}

	if err != nil {
		log.Printf("Failed deleting %v", err)
		return err
	}
	return nil
}

// ReplaceAll first delete all and then create many sets in a mongo db
func (collection *SetCollection) ReplaceAll(sets []*Set) (*[]primitive.ObjectID, error) {
	ctx := collection.Context

	err := collection.Collection.Drop(ctx)
	if err != nil {
		return nil, err
	}

	var ui []interface{}
	for _, t := range sets {
		t.ID = primitive.NewObjectID()
		ui = append(ui, t)
	}

	result, err := collection.Collection.InsertMany(ctx, ui)
	if err != nil {
		log.Printf("Could not create Set: %v", err)
		return nil, err
	}

	var oids = []primitive.ObjectID{}

	for _, id := range result.InsertedIDs {
		oids = append(oids, id.(primitive.ObjectID))
	}
	return &oids, nil
}
