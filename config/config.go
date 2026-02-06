package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	mongoURI, exists := os.LookupEnv("MONGO_URI")
	if !exists {
		log.Fatal("MONGO_URI environment variable not set")
	}

	return &Config{
		MongoURI: mongoURI,
	}
}
