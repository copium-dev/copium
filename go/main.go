package main

import (
	"fmt"
	"jtracker-backend/config"
	"jtracker-backend/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// googleAccessKey := os.Getenv("GOOGLE_CLIENT_ID")
	// googleSecretKey := os.Getenv("GOOGLE_CLIENT_SECRET")

	config.LoadConfig()

	// Initialize the router
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HomeHandler)
	r.HandleFunc("/login", handlers.LoginHandler)
	r.HandleFunc("/auth/callback", handlers.CallbackHandler)

	port := ":8080"
	fmt.Println("Server started at http://localhost" + port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
