package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI string
	JWTKey   []byte
}

func Load() *Config {
	// Try loading from .env for local development
	_ = godotenv.Load()

	mongoURI, exists := os.LookupEnv("MONGO_URI")
	jwtSecret, jwtExists := os.LookupEnv("JWT_SECRET")
	if !jwtExists {
		log.Fatal("JWT_SECRET environment variable not set")
	}
	jwtKey := []byte(jwtSecret)
	if !exists {
		log.Fatal("MONGO_URI environment variable not set")
	}

	log.Printf("Loaded MONGO_URI: %s", maskURI(mongoURI))

	return &Config{
		MongoURI: mongoURI,
		JWTKey:   jwtKey,
	}
}

// maskURI hides sensitive info for logging
func maskURI(uri string) string {
	if len(uri) > 30 {
		return uri[:10] + "..." + uri[len(uri)-10:]
	}
	return uri
}
