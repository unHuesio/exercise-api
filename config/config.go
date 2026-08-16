package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI        string
	JWTKey          []byte
	GoogleClientID  string
	GoogleIssuer    string
	ExercisesFile   string
	ExercisesBucket string
	ExercisesObject string
}

func Load() *Config {
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

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleIssuer := os.Getenv("GOOGLE_ISSUER")
	if googleIssuer == "" {
		googleIssuer = "https://accounts.google.com"
	}

	exercisesFile := os.Getenv("EXERCISES_FILE")
	if exercisesFile == "" {
		exercisesFile = "excercises.json"
	}

	exercisesBucket := os.Getenv("EXERCISES_BUCKET")
	exercisesObject := os.Getenv("EXERCISES_OBJECT")
	if exercisesObject == "" {
		exercisesObject = "excercises.json"
	}

	log.Printf("Loaded MONGO_URI: %s", maskURI(mongoURI))

	return &Config{
		MongoURI:        mongoURI,
		JWTKey:          jwtKey,
		GoogleClientID:  googleClientID,
		GoogleIssuer:    googleIssuer,
		ExercisesFile:   exercisesFile,
		ExercisesBucket: exercisesBucket,
		ExercisesObject: exercisesObject,
	}
}

// maskURI hides sensitive info for logging
func maskURI(uri string) string {
	if len(uri) > 30 {
		return uri[:10] + "..." + uri[len(uri)-10:]
	}
	return uri
}
