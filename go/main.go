package main

import (
    "log"
    "net/http"
    "os"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "github.com/joho/godotenv"
)

var googleOauthConfig *oauth2.Config

func main() {
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }

    googleOauthConfig = &oauth2.Config{
        RedirectURL:  "http://localhost:8080/auth",
        ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
        ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
        Scopes: []string{
            "https://www.googleapis.com/auth/userinfo.email",
            "https://www.googleapis.com/auth/userinfo.profile",
        },
        Endpoint: google.Endpoint,
    }

    server := &http.Server{
        Addr:    ":8080",
        Handler: New(),
    }

    log.Printf("Starting HTTP Server on :8080")
    if err := server.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatal(err)
    }
}

func New() http.Handler {
    mux := http.NewServeMux()
    
    // CORS middleware
    corsMiddleware := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
            w.Header().Set("Access-Control-Allow-Credentials", "true")
            
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }

    // Routes
    mux.HandleFunc("/auth", oauthGoogleLogin)
   // mux.HandleFunc("/auth/google/callback", oauthGoogleCallback)

    return corsMiddleware(mux)
}

func oauthGoogleLogin(w http.ResponseWriter, r *http.Request) {
    url := googleOauthConfig.AuthCodeURL("state")
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func oauthGoogleCallback(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement callback handling
    // Exchange code for token
    // Get user info
    // Create session
    // Set cookie
    // Redirect to frontend
}