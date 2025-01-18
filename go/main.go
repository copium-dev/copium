package main

import (
	"jtracker-backend/handlers"
	// "log"
	"net/http"
	// "os"

	"github.com/gorilla/mux"
	// "github.com/joho/godotenv"
)

func main() {

	// Initialize the router
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HomeHandler)
	r.HandleFunc("/login", handlers.LoginHandler)
	r.HandleFunc("/auth/callback", handlers.CallbackHandler)

	http.ListenAndServe(":8080", r)
}
