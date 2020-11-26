package db

import (
	"context"
	"errors"
	"log"

	"bitbucket.org/spinnerweb/cards/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GameType string
type LegalText string
type Currency string
type RelatedURI string

const (
	Large     string    = "large"
	Normal    string    = "normal"
	Commander GameType  = "commander"
	Legal     LegalText = "legal"
	USD       Currency  = "usd"
)

type ScryfallCard struct {
	ID            string                 `bson:"_id" json:"id"`
	OracleID      string                 `json:"oracle_id"`
	Name          string                 `json:"name"`
	Lang          string                 `json:"lang"`
	ScryfallURI   string                 `json:"scryfall_uri"`
	HighresImage  bool                   `json:"highres_image"`
	Layout        string                 `json:"layout"`
	ImageURLs     map[string]string      `json:"image_uris"`
	ManaCost      string                 `json:"mana_cost"`
	Cmc           float64                `json:"cmc"`
	TypeLine      string                 `json:"type_line"`
	OracleText    string                 `json:"oracle_text"`
	Colors        []string               `json:"colors"`
	ColorIdentity []string               `json:"color_identity"`
	Power         string                 `json:"power"`
	Toughness     string                 `json:"toughness"`
	Keywords      []string               `json:"keywords"`
	CardFaces     []ScryfallCard         `json:"card_faces"`
	Legalities    map[GameType]LegalText `json:"legalities"`
	Set           string                 `json:"set"`
	SetName       string                 `json:"set_name"`
	SetType       string                 `json:"set_type"`
	SetURI        string                 `json:"scryfall_set_uri"`
	RulingsURI    string                 `json:"rulings_uri"`
	Rarity        string                 `json:"rarity"`
	EdhrecRank    int                    `json:"edhrec_rank"`
	Prices        map[Currency]string    `json:"prices"`
	RelatedURIs   map[RelatedURI]string  `json:"related_uris"`
}

// ScryfallCardCollection ...
type ScryfallCardCollection struct {
	*mongo.Client
	*mongo.Collection
	context.Context
	context.CancelFunc
}

// GetScryfallCardCollection ...
func GetScryfallCardCollection() ScryfallCardCollection {
	client, ctx, cancel := db.GetConnection()
	db := client.Database(db.GetDatabaseName())
	collection := db.Collection("scryfallcards")

	return ScryfallCardCollection{
		Client:     client,
		Collection: collection,
		Context:    ctx,
		CancelFunc: cancel,
	}
}

// Disconnect ...
func (collection *ScryfallCardCollection) Disconnect() {
	collection.CancelFunc()
	collection.Client.Disconnect(collection.Context)
}

// GetAllScryfallCards Retrives all scryfallcards from the db
func (collection *ScryfallCardCollection) GetAllScryfallCards() ([]*ScryfallCard, error) {
	var scryfallcards []*ScryfallCard = []*ScryfallCard{}
	ctx := collection.Context

	cursor, err := collection.Collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &scryfallcards)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return scryfallcards, nil
}

// GetScryfallCardByID Retrives a scryfallcard by its id from the db
func (collection *ScryfallCardCollection) GetScryfallCardByID(id string) (*ScryfallCard, error) {
	var scryfallcard *ScryfallCard
	ctx := collection.Context

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Failed parsing id %v", err)
		return nil, err
	}
	result := collection.Collection.FindOne(ctx, bson.D{bson.E{Key: "_id", Value: objID}})
	if result == nil {
		return nil, errors.New("Could not find a ScryfallCard")
	}
	err = result.Decode(&scryfallcard)

	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return scryfallcard, nil
}

// GetScryfallCardByIDs Retrives scryfallcards by their ids from the db
func (collection *ScryfallCardCollection) GetScryfallCardByIDs(ids *[]primitive.ObjectID) (*[]ScryfallCard, error) {
	ctx := collection.Context
	var scryfallcards []ScryfallCard = []ScryfallCard{}
	cursor, err := collection.Collection.Find(ctx, bson.D{bson.E{Key: "_id", Value: bson.M{"$in": ids}}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &scryfallcards)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return &scryfallcards, nil

}

// GetScryfallCardByKey Retrives a scryfallcard by its key from the db
func (collection *ScryfallCardCollection) GetScryfallCardByKey(key string) (*ScryfallCard, error) {
	var scryfallcard *ScryfallCard
	ctx := collection.Context

	result := collection.Collection.FindOne(ctx, bson.D{bson.E{Key: "key", Value: key}})
	if result == nil {
		return nil, nil
	}
	err := result.Decode(&scryfallcard)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return scryfallcard, nil
}

// Create creating a scryfallcard in a mongo
func (collection *ScryfallCardCollection) Create(scryfallcard *ScryfallCard) (primitive.ObjectID, error) {
	ctx := collection.Context
	scryfallcard.ID = primitive.NewObjectID().Hex()

	result, err := collection.Collection.InsertOne(ctx, scryfallcard)
	if err != nil {
		log.Printf("Could not create ScryfallCard: %v", err)
		return primitive.NilObjectID, err
	}
	oid := result.InsertedID.(primitive.ObjectID)
	return oid, nil
}

// CreateMany creating many scryfallcards in a mongo
func (collection *ScryfallCardCollection) CreateMany(scryfallcards []*ScryfallCard) (*[]string, error) {
	ctx := collection.Context

	var ui []interface{}
	for _, t := range scryfallcards {
		if t.ID == "" {
			t.ID = primitive.NewObjectID().Hex()
		}
		ui = append(ui, t)
	}

	result, err := collection.Collection.InsertMany(ctx, ui)
	if err != nil {
		log.Printf("Could not create ScryfallCard: %v", err)
		return nil, err
	}

	var oids = []string{}

	for _, id := range result.InsertedIDs {
		oids = append(oids, id.(string))
	}
	return &oids, nil
}

//Update updating an existing scryfallcard in a mongo
func (collection *ScryfallCardCollection) Update(scryfallcard *ScryfallCard) (*ScryfallCard, error) {
	ctx := collection.Context
	var updatedScryfallCard *ScryfallCard

	update := bson.M{
		"$set": scryfallcard,
	}

	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		Upsert:         &upsert,
		ReturnDocument: &after,
	}

	err := collection.Collection.FindOneAndUpdate(ctx, bson.M{"_id": scryfallcard.ID}, update, &opt).Decode(&updatedScryfallCard)
	if err != nil {
		log.Printf("Could not save ScryfallCard: %v", err)
		return nil, err
	}
	return updatedScryfallCard, nil
}

// DeleteScryfallCardByID Deletes an scryfallcard by its id from the db
func (collection *ScryfallCardCollection) DeleteScryfallCardByID(id string) error {
	ctx := collection.Context
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Failed parsing id %v", err)
		return err
	}
	result, err := collection.Collection.DeleteOne(ctx, bson.D{bson.E{Key: "_id", Value: objID}})
	if result == nil {
		return errors.New("Could not find a ScryfallCard")
	}

	if err != nil {
		log.Printf("Failed deleting %v", err)
		return err
	}
	return nil
}

// ReplaceAll first delete all and then create many cards in a mongo db
func (collection *ScryfallCardCollection) ReplaceAll(cards []*ScryfallCard) (*[]string, error) {
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
