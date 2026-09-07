package main

import (
	"context"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"traffic-backend/internal/config"
	"traffic-backend/internal/handler"
	"traffic-backend/internal/repository"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	ctx := context.Background()

	db, err := config.ConnectDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	trafficRepo := repository.NewTrafficRepository(db)

	// Automatic cleanup of traffic data older than 3 days.
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			if err := trafficRepo.DeleteOldTrafficData(context.Background()); err != nil {
				log.Println("Traffic cleanup ERROR:", err)
			} else {
				log.Println("Old traffic data cleanup completed")
			}
		}
	}()

	trafficHandler := handler.NewTrafficHandler(trafficRepo)

	http.HandleFunc(
		"/api/v1/traffic/events",
		trafficHandler.HandleTrafficEvent,
	)

	http.HandleFunc(
		"/api/v1/traffic/current",
		trafficHandler.HandleCurrentTraffic,
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf(
		"Server running on http://localhost:%s",
		port,
	)

	handler := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Content-Type",
		},
	}).Handler(http.DefaultServeMux)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
