package main

import (
	"context"
	"log"

	"github.com/INOVA/DML/internal/config"
	"github.com/INOVA/DML/internal/db"
	internalHTTP "github.com/INOVA/DML/internal/http"
)

// @title           DML API
// @version         1.0
// @description     Deep QMS backend application serving the core HR, IT, and specialized logic domains.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8081
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 1. Load configuration (from .env or environment variables)
	cfg := config.Load()

	// 2. Initialize Database Connection
	database, err := db.New(context.Background(), cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// 3. Initialize and Run HTTP Server
	server := internalHTTP.NewServer(cfg, database)

	log.Println("Successfully initialized dependencies. Starting application...")
	if err := server.Run(); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
