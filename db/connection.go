package db

import (
	"flag"
	"time"

	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// Timeout operations after N seconds
	connectTimeout = 60
)

// GetConnection Creates and returns the connection
func GetConnection() (*mongo.Client, context.Context, context.CancelFunc) {
	protocol := os.Getenv("MONGODB_PROTOCOL")
	username := os.Getenv("MONGODB_USERNAME")
	password := os.Getenv("MONGODB_PASSWORD")
	clusterEndpoint := os.Getenv("MONGODB_ENDPOINT")
	log.Printf("Connection to %s\n", clusterEndpoint)

	connectionURI := fmt.Sprintf("%s://%s:%s@%s", protocol, username, password, clusterEndpoint)

	client, err := mongo.NewClient(options.Client().ApplyURI(connectionURI))
	if err != nil {
		log.Printf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout*time.Second)

	err = client.Connect(ctx)
	if err != nil {
		log.Printf("Failed to connect to cluster: %v", err)
	}

	// Force a connection to verify our connection string
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("Failed to ping cluster: %v", err)
	}

	fmt.Println("Connected to MongoDB!")
	return client, ctx, cancel
}

// GetDatabaseName ...
func GetDatabaseName() string {
	if flag.Lookup("test.v") == nil {
		fmt.Println("normal run")
		return "mtg"
	}
	fmt.Println("run under go test")
	return "mtg_test"
}
