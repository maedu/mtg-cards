package db

import (
	"context"
	"log"

	"github.com/maedu/mtg-cards/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id" json:"-"`
	UserID   string             `bson:"user_id" json:"userId"`
	UserName string             `bson:"user_name" json:"userName"`
}

// UserCollection ...
type UserCollection struct {
	*mongo.Client
	*mongo.Collection
	context.Context
	context.CancelFunc
}

// GetUserCollection ...
func GetUserCollection() (UserCollection, error) {
	client, ctx, cancel := db.GetConnection()
	db := client.Database(db.GetDatabaseName())
	collection := db.Collection("users")

	model := mongo.IndexModel{
		Keys: bson.M{
			"user_id": 1,
		}, Options: nil,
	}
	_, err := collection.Indexes().CreateOne(ctx, model)

	return UserCollection{
		Client:     client,
		Collection: collection,
		Context:    ctx,
		CancelFunc: cancel,
	}, err
}

// Disconnect ...
func (collection *UserCollection) Disconnect() {
	collection.CancelFunc()
	collection.Client.Disconnect(collection.Context)
}

// GetAllUsers Retrives all users from the db
func (collection *UserCollection) GetAllUsers() ([]*User, error) {
	var users []*User = []*User{}
	ctx := collection.Context

	cursor, err := collection.Collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &users)
	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return users, nil
}

// GetUserByUserID retrives a user by its user id (aka email)
func (collection *UserCollection) GetUserByUserID(userID string) (*User, error) {
	var user *User
	ctx := collection.Context

	result := collection.Collection.FindOne(ctx, bson.D{bson.E{Key: "user_id", Value: userID}})
	if result == nil {
		return nil, nil
	}
	err := result.Decode(&user)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return user, nil
}

// GetUserByUserID retrives a user by its user name
func (collection *UserCollection) GetUserByUserName(userName string) (*User, error) {
	var user *User
	ctx := collection.Context

	result := collection.Collection.FindOne(ctx, bson.D{bson.E{Key: "user_name", Value: userName}})
	if result == nil {
		return nil, nil
	}
	err := result.Decode(&user)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		log.Printf("Failed marshalling %v", err)
		return nil, err
	}
	return user, nil
}

// Create creating a user in a mongo
func (collection *UserCollection) Create(user *User) (primitive.ObjectID, error) {
	ctx := collection.Context
	user.ID = primitive.NewObjectID()

	result, err := collection.Collection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("Could not create User: %v", err)
		return primitive.NilObjectID, err
	}
	oid := result.InsertedID.(primitive.ObjectID)
	return oid, nil
}

func (collection *UserCollection) Update(user *User) (*User, error) {
	ctx := collection.Context
	var updatedUser *User

	update := bson.M{
		"$set": user,
	}

	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		Upsert:         &upsert,
		ReturnDocument: &after,
	}

	err := collection.Collection.FindOneAndUpdate(ctx, bson.M{"_id": user.ID}, update, &opt).Decode(&updatedUser)
	if err != nil {
		log.Printf("Could not save User: %v", err)
		return nil, err
	}
	return updatedUser, nil
}
