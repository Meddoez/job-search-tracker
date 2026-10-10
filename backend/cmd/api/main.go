package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	database "github.com/Meddoez/job-search-tracker/backend/internal/database"

	"github.com/Meddoez/job-search-tracker/backend/internal/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	db, err := database.ConnectDatabase()
	if err != nil {
		fmt.Println("Failed to connect to the database:", err)
		panic(err)
	}

	defer db.Close()

	ingestionServiceURL := strings.TrimSpace(os.Getenv("INGESTION_SERVICE_URL"))
	if ingestionServiceURL == "" {
		log.Fatal("INGESTION_SERVICE_URL is required")
	}
	client := &http.Client{Timeout: 90 * time.Second}
	profileRepo := database.NewProfileRepository(db)
	profileJobRepo := database.NewProfileJobRepository(db)

	routes.RegisterRoutes(router, client, ingestionServiceURL, profileRepo, profileJobRepo)
	router.Run()
}
