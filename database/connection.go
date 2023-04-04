package database

import (
	"context"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)


func Connect() {

	envError := godotenv.Load()

	if envError != nil {
		log.Fatal(".env file couldn't be loaded")
	}

	mongoURL := os.Getenv("MONGO__URL")

	clientOptions := options.Client().ApplyURI(mongoURL)
	_, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to DB")
}
