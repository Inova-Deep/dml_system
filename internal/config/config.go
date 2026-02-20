package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	APIPort     string
	DBDSN       string
	CORSOrigins []string
	JWTSecret   string
}

// Load loads environment variables into the Config struct.
func Load() *Config {
	// Attempt to load .env, ignore if it doesn't exist (e.g., in production)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found; assuming environment variables are set")
	}

	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8081" // fallback default
	}

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		log.Fatal("DB_DSN environment variable is strictly required")
	}

	corsOriginsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	var corsOrigins []string
	if corsOriginsEnv == "" {
		// Default to allowing local development environments if none set
		corsOrigins = []string{"http://localhost:3000", "http://localhost:5173"}
	} else {
		corsOrigins = strings.Split(corsOriginsEnv, ",")
		for i := range corsOrigins {
			corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
		}
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// VERY bad practice for prod, but acceptable warning for local scaffolding without it
		log.Println("WARNING: JWT_SECRET environment variable is not set. Using insecure default secret!")
		jwtSecret = "super-insecure-local-dev-secret-do-not-use"
	}

	return &Config{
		APIPort:     apiPort,
		DBDSN:       dbDSN,
		CORSOrigins: corsOrigins,
		JWTSecret:   jwtSecret,
	}
}
