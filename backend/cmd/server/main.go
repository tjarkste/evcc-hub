package main

import (
	"log"
	"os"

	"evcc-cloud/backend/internal/api"
	"evcc-cloud/backend/internal/storage"
)

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data.db"
	}

	env := os.Getenv("ENV")
	devMode := env == "development"

	db, err := storage.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := api.NewRouter(db, api.Config{
		JWTSecret: jwtSecret,
		DevMode:   devMode,
	})

	log.Printf("evcc-cloud backend starting on :%s (dev=%v)", port, devMode)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
