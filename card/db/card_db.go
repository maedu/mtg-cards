package db

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	pagination "github.com/maedu/mongo-go-pagination"
	"github.com/maedu/mtg-cards/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type Set string
type Rarity string

type PaginatedResult struct {
	Cards      []*Card                   `json:"cards"`
	Pagination pagination.PaginationData `json:"pagination"`
}

type CardType string

const (
	Artifact     CardType = "Artifact"
	Conspiracy   CardType = "Conspiracy"
	Creature     CardType = "Creature"
	Enchantment  CardType = "Enchantment"
	Instant      CardType = "Instant"
	Land         CardType = "Land"
	Phenomenon   CardType = "Phenomenon"
	Plane        CardType = "Plane"
	Planeswalker CardType = "Planeswalker"
	Scheme       CardType = "Scheme"
	Sorcery      CardType = "Sorcery"
	Tribal       CardType = "Tribal"
	Vanguard     CardType = "Vanguard"

	PriceFilterSkipped float64 = -10
)

type Card struct {
	ID              primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	ScryfallID      string             `bson:"scryfall_id" json:"-"`
	Name            string             `json:"name"`
	Lang            string             `json:"lang"`
	Layout          string             `json:"layout"`
	ImageURLs       map[string]string  `json:"imageURLs"`
	ManaCost        string             `json:"manaCost"`
	Cmc             float64            `json:"cmc"`
	TypeLine        string             `bson:"type_line" json:"typeLine"`
	CardTypes       []CardType         `bson:"card_types" json:"cardTypes"`
	OracleText      string             `bson:"oracle_text" json:"-"`
	Colors          []string           `json:"colors"`
	ColorIdentity   []string           `json:"colorIdentity"`
	Keywords        []string           `json:"keywords"`
	LegalInComander bool               `json:"legalInCommander"`
	SetName         string             `bson:"set_name" json:"setName"`
	Rarity          string             `json:"rarity"`
	EdhrecRank      int                `json:"edhrecRank"`
	Price           float64            `json:"price"`
	Score           float64            `bson:"score" json:"score"`
	CardFaces       []Card             `bson:"card_faces" json:"cardFaces"`
	IsCommander     bool               `bson:"is_commander" json:"isCommander"`
	IsLand          bool               `bson:"is_land" json:"-"`
	SearchText      string             `bson:"search_text" json:"searchText"`
	CardGroups      []string           `bson:"card_groups" json:"cardGroups"`
	Synergies       map[string]float64 `bson:"synergies" json:"synergies"`
	InCollection    bool               `bson:"-" json:"inCollection"`
}

type CardSearchRequest struct {
	Text                    string
	Cmc                     []float64
	Colors                  []string
	CardGroups              []string
	MainCardForSynergy      string
	SearchRelatedToMainCard bool
	PriceMin                float64
	PriceMax                float64
	SortBy                  string
	SortDir                 string
	UserID                  string
}

// CardCollection ...
type CardCollection struct {
	*mongo.Client
	*mongo.Collection
	context.Context
	context.CancelFunc
}

// GetCardCollection ...
func GetCardCollection() (CardCollection, error) {
	client, ctx, cancel := db.GetConnection()
	db := client.Database(db.GetDatabaseName())
	collection := db.Collection("cards")
	modelOpts := options.Index()
	modelOpts.SetWeights(bson.M{
		"name":        5,
		"oracle_text": 4,
		"type_line":   2,
		"card_groups": 2,
		"set_name":    1,
	})
	index := mongo.IndexModel{
		Keys: bsonx.Doc{
			{Key: "name", Value: bsonx.String("text")},
			{Key: "oracle_text", Value: bsonx.String("text")},
			{Key: "type_line", Value: bsonx.String("text")},
			{Key: "card_groups", Value: bsonx.String("text")},
			{Key: "set_name", Value: bsonx.String("text")},
		},
		Options: modelOpts,
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	indexName, err := collection.Indexes().CreateOne(ctx, index, opts)
	if err != nil {
		return CardCollection{}, err
	}
	fmt.Printf("Index created: %s\n", indexName)

	return CardCollection{
		Client:     client,
		Collection: collection,
		Context:    ctx,
		CancelFunc: cancel,
	}, nil
}

// Disconnect ...
func (collection *CardCollection) Disconnect() {
	collection.CancelFunc()
	collection.Client.Disconnect(collection.Context)
}

// GetAllCards Retrives all cards from the db
func (collection *CardCollection) GetAllCards() ([]*Card, error) {
	var cards []*Card = []*Card{}
	ctx := collection.Context

	cursor, err := collection.Collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &cards)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return cards, nil
}

func getFilter(request CardSearchRequest) (bson.M, bson.M) {

	projection := bson.M{}
	trimmedText := strings.TrimSpace(request.Text)
	filters := []bson.M{}
	if trimmedText != "" {
		filters = append(filters, bson.M{"$text": bson.M{
			"$search": trimmedText,
		}})
		projection = bson.M{
			"score": bson.M{
				"$meta": "textScore",
			},
		}

	}

	if request.Cmc != nil && len(request.Cmc) > 0 {
		cmcFilter := []bson.M{}
		for _, cmc := range request.Cmc {
			comparison := "$eq"
			if cmc <= 1 {
				comparison = "$lte"
			} else if cmc >= 7 {
				comparison = "$gte"
			}
			cmcFilter = append(cmcFilter, bson.M{
				"cmc": bson.M{comparison: cmc},
			})
		}

		filters = append(filters, bson.M{"$or": cmcFilter})
	}

	if request.Colors != nil && len(request.Colors) > 0 {
		filteredColors := []string{}
		muliColorSelected := false
		for _, color := range request.Colors {
			if color == "M" {
				muliColorSelected = true
			} else {
				filteredColors = append(filteredColors, color)
			}
		}
		sort.Strings(filteredColors)

		if muliColorSelected {
			filters = append(filters, bson.M{"colors.1": bson.M{"$exists": true}})

			if len(filteredColors) > 0 {
				// Include cards which match exactly the selected colors
				filters = append(filters, bson.M{"colors": bson.M{"$eq": filteredColors}})
			}

		} else {
			// Include cards have at least one of the selected colors, but no others.
			filters = append(filters, bson.M{"colors": bson.M{"$not": bson.M{"$elemMatch": bson.M{
				"$nin": filteredColors,
			}}}})
		}
	}

	if request.CardGroups != nil && len(request.CardGroups) > 0 {
		cardGroups := []string{}
		synergyFound := false
		for _, cardGroup := range request.CardGroups {
			if cardGroup == "Synergy" {
				synergyFound = true
			} else if cardGroup != "Collected" {
				cardGroups = append(cardGroups, cardGroup)
			}
		}

		if len(cardGroups) > 0 {
			filters = append(filters, bson.M{"card_groups": bson.M{
				"$all": cardGroups,
			}})
		}

		if synergyFound {
			// Filter for synergy
			filters = append(filters, bson.M{"synergies." + request.MainCardForSynergy: bson.M{
				"$gte": 0.2,
			}})
		}
	}

	if request.SearchRelatedToMainCard {
		filters = append(filters, bson.M{"synergies." + request.MainCardForSynergy: bson.M{
			"$exists": true,
		}})
	}

	if request.PriceMin > PriceFilterSkipped {
		filters = append(filters, bson.M{"price": bson.M{
			"$gte": request.PriceMin,
		}})
	}

	if request.PriceMax > PriceFilterSkipped {
		filters = append(filters, bson.M{"price": bson.M{
			"$lte": request.PriceMax,
		}})
	}

	filter := bson.M{}
	if len(filters) > 0 {
		filter = bson.M{
			"$and": filters,
		}
	}

	fmt.Printf("\nFilter: %v\n", filter)

	return filter, projection

}

// GetCardsPaginated Retrives all cards from the db
func (collection *CardCollection) GetCardsPaginated(limit int64, page int64, request CardSearchRequest) (PaginatedResult, error) {
	var cards []*Card = []*Card{}

	filter, projection := getFilter(request)
	sort := getSortOptions(request)
	paginatedData, err := pagination.New(collection.Collection).Limit(limit).Page(page).Filter(filter).Select(projection).Sort(sort).Find()
	if err != nil {
		return PaginatedResult{}, err
	}

	for _, raw := range paginatedData.Data {
		var card *Card
		if marshallErr := bson.Unmarshal(raw, &card); marshallErr != nil {
			log.Printf("Failed marshalling: %v", err)
			return PaginatedResult{}, err
		}
		cards = append(cards, card)

	}
	return PaginatedResult{
		Cards:      cards,
		Pagination: paginatedData.Pagination,
	}, nil
}

// GetCollectedCardsPaginated Retrives all cards from the db
func (collection *CardCollection) GetCollectedCardsPaginated(limit int64, page int64, request CardSearchRequest) (PaginatedResult, error) {
	var cards []*Card = []*Card{}

	fmt.Println("GetCollectedCardsPaginated")

	filter, projection := getFilter(request)

	matchStage := bson.M{"$match": filter}

	lookupUserCards := bson.M{"$lookup": bson.D{
		primitive.E{Key: "from", Value: "user_cards"},
		primitive.E{Key: "localField", Value: "name"},
		primitive.E{Key: "foreignField", Value: "name"},
		primitive.E{Key: "as", Value: "user_cards"},
	},
	}
	fmt.Println(lookupUserCards)

	matchForUserStage := bson.M{"$match": bson.M{"user_cards.user_id": bson.M{"$eq": request.UserID}}}

	sort := getSortOptions(request)

	paginatedData, err := pagination.New(collection.Collection).Limit(limit).Page(page).Select(projection).Sort(sort).Aggregate(
		matchStage,
		lookupUserCards,
		matchForUserStage,
	)
	if err != nil {
		return PaginatedResult{}, err
	}

	for _, raw := range paginatedData.Data {
		var card *Card
		if marshallErr := bson.Unmarshal(raw, &card); marshallErr != nil {
			log.Printf("Failed marshalling: %v", err)
			return PaginatedResult{}, err
		}
		cards = append(cards, card)

	}
	return PaginatedResult{
		Cards:      cards,
		Pagination: paginatedData.Pagination,
	}, nil
}

func getSortOptions(request CardSearchRequest) bson.D {
	trimmedText := strings.TrimSpace(request.Text)
	if trimmedText != "" {
		return bson.D{
			primitive.E{Key: "score", Value: bson.M{
				"$meta": "textScore",
			}}}
	}

	sortDir := 1
	if request.SortDir == "desc" {
		sortDir = -1
	}
	sortBy := request.SortBy
	if sortBy == "synergy" {
		sortBy = "synergies." + request.MainCardForSynergy
	}

	sort := bson.D{
		primitive.E{Key: "is_land", Value: 1},
		primitive.E{Key: sortBy, Value: sortDir},
	}

	return sort
}

// GetCardsByNames retrieves cards by their names from the db
func (collection *CardCollection) GetCardsByNames(names []string) ([]*Card, error) {
	log.Printf("find cards by names: %d", len(names))
	var cards []*Card = []*Card{}
	ctx := collection.Context

	inFilter := bson.M{"$in": names}

	filter := bson.M{"$or": bson.A{
		bson.M{"name": inFilter},
		bson.M{"card_faces.name": inFilter},
	}}

	cursor, err := collection.Collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("find failed: %w", err)
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &cards)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return cards, nil
}

// GetCardsBySetName retrieves a card by its set name from the db
func (collection *CardCollection) GetCardsBySetName(setName string) ([]*Card, error) {
	var cards []*Card = []*Card{}
	ctx := collection.Context

	cursor, err := collection.Collection.Find(ctx, bson.D{bson.E{Key: "set_name", Value: setName}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &cards)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return cards, nil
}

// Create creating a card in a mongo
func (collection *CardCollection) Create(card *Card) (primitive.ObjectID, error) {
	ctx := collection.Context
	card.ID = primitive.NewObjectID()

	result, err := collection.Collection.InsertOne(ctx, card)
	if err != nil {
		log.Printf("Could not create Card: %v", err)
		return primitive.NilObjectID, err
	}
	oid := result.InsertedID.(primitive.ObjectID)
	return oid, nil
}

// DeleteAll cards in the collection
func (collection *CardCollection) DeleteAll() error {
	ctx := collection.Context

	return collection.Collection.Drop(ctx)
}

// CreateMany creating many cards in a mongo
func (collection *CardCollection) CreateMany(cards []*Card) error {
	ctx := collection.Context

	var ui []interface{}
	for _, t := range cards {
		t.ID = primitive.NewObjectID()
		ui = append(ui, t)
	}

	_, err := collection.Collection.InsertMany(ctx, ui)
	if err != nil {
		log.Printf("Could not create Card: %v", err)
		return err
	}

	return nil
}

// ReplaceAll first delete all and then create many cards in a mongo db
func (collection *CardCollection) ReplaceAll(cards []*Card) (*[]primitive.ObjectID, error) {
	ctx := collection.Context

	err := collection.Collection.Drop(ctx)
	if err != nil {
		return nil, err
	}

	var ui []interface{}
	for _, t := range cards {
		t.ID = primitive.NewObjectID()
		ui = append(ui, t)
	}

	result, err := collection.Collection.InsertMany(ctx, ui)
	if err != nil {
		log.Printf("Could not create Card: %v", err)
		return nil, err
	}

	var oids = []primitive.ObjectID{}

	for _, id := range result.InsertedIDs {
		oids = append(oids, id.(primitive.ObjectID))
	}
	return &oids, nil
}
