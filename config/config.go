package config

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI string
	JWTKey   []byte
	// GoogleClientID is the primary (e.g. web) OAuth client ID, kept for
	// backward compatibility. GoogleClientIDs holds every accepted audience
	// (web, Android, iOS, ...) and should be used for token verification.
	GoogleClientID  string
	GoogleClientIDs []string
	GoogleIssuer    string
	ExercisesFile   string
	ExercisesBucket string
	ExercisesObject string
	AllowedOrigins  []string
}

func Load() *Config {
	_ = godotenv.Load()

	mongoURI, exists := os.LookupEnv("MONGO_URI")
	jwtSecret, jwtExists := os.LookupEnv("JWT_SECRET")
	if !jwtExists {
		log.Fatal("JWT_SECRET environment variable not set")
	}
	jwtSecret = strings.TrimSpace(jwtSecret)
	// An empty or too-short secret would let anyone forge valid HS256 JWTs.
	if len(jwtSecret) < 8 {
		log.Fatal("JWT_SECRET must be set to a non-empty value of at least 8 characters")
	}
	jwtKey := []byte(jwtSecret)
	if !exists {
		log.Fatal("MONGO_URI environment variable not set")
	}

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	// GOOGLE_CLIENT_ID and GOOGLE_IOS_CLIENT_ID may each hold a comma-separated
	// list, allowing multiple OAuth clients (e.g. web + iOS) to log in.
	googleClientIDs := parseCommaSeparated(googleClientID)
	googleClientIDs = append(googleClientIDs, parseCommaSeparated(os.Getenv("GOOGLE_IOS_CLIENT_ID"))...)
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
		GoogleClientIDs: googleClientIDs,
		GoogleIssuer:    googleIssuer,
		ExercisesFile:   exercisesFile,
		ExercisesBucket: exercisesBucket,
		ExercisesObject: exercisesObject,
		AllowedOrigins:  parseCommaSeparated(os.Getenv("ALLOWED_ORIGINS")),
	}
}

// parseCommaSeparated parses a comma-separated list of values (origins,
// OAuth client IDs, ...), trimming whitespace and dropping empty entries.
// An empty/unset value returns nil.
func parseCommaSeparated(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

// maskURI hides sensitive info (e.g. embedded credentials) for logging.
func maskURI(uri string) string {
	parsed, err := url.Parse(uri)
	if err != nil || parsed.Host == "" {
		// Not a parseable URI (or no host); avoid logging it verbatim since it
		// may still contain a password, regardless of its length.
		return "****"
	}
	parsed.User = nil
	return parsed.String()
}
