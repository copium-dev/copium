package config

import (
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// var GoogleOauthConfig = &oauth2.Config{
// 	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
// 	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
// 	RedirectURL:  "http://localhost:8080/auth/callback",
// 	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
// 	Endpoint:     google.Endpoint,
// }

var GoogleOauthConfig *oauth2.Config

func LoadConfig() {
	// Fetch the environment variables directly
	googleAccessKey := os.Getenv("GOOGLE_CLIENT_ID")
	googleSecretKey := os.Getenv("GOOGLE_CLIENT_SECRET")

	// Log the values (for debugging)
	log.Println("GOOGLE_CLIENT_ID:", googleAccessKey)
	log.Println("GOOGLE_CLIENT_SECRET:", googleSecretKey)

	// If any of the credentials are missing, log a fatal error and exit
	if googleAccessKey == "" || googleSecretKey == "" {
		log.Fatal("Missing Google OAuth client ID or client secret in environment variables.")
	}

	// Set up the OAuth2 config with the loaded credentials
	GoogleOauthConfig = &oauth2.Config{
		ClientID:     googleAccessKey,
		ClientSecret: googleSecretKey,
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
}
