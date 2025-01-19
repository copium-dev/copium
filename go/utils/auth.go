package utils

import (
    "log"
    "sync"
    "net/http"
    "os"

    "github.com/gorilla/sessions"
    "github.com/joho/godotenv"
    "github.com/markbates/goth"
    "github.com/markbates/goth/gothic"
    "github.com/markbates/goth/providers/google"
)

type AuthHandler struct {
    GoogleClientID     string
    GoogleClientSecret string
    CallbackUrl        string
    JwtSecret          string
    Store              *sessions.CookieStore
}

var (
    store *sessions.CookieStore
    once sync.Once
)

// we only want to initialize the AuthHandler ONCE
// previously, the Store kept getting re-initialized because 
// the nature of gorilla/mux is that it spawns a new goroutine for each request
// as such, each goroutine could potentially create its own handler and store
// and overwrite the global gothic.Store (we want to use the same store for all requests)
func NewAuthHandler() *AuthHandler {
    once.Do(func() {
        err := godotenv.Load()
        if err != nil {
            log.Fatal("Error loading .env file")
        }

        googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
        googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
        callbackUrl := os.Getenv("CALLBACK_URL")
        jwtSecret := os.Getenv("JWT_SECRET")

        store = sessions.NewCookieStore([]byte(jwtSecret))
        store.Options = &sessions.Options{
            Path:     "/",
            MaxAge:   86400 * 30,
            HttpOnly: true,
            Secure:   false,
            SameSite: http.SameSiteLaxMode,
        }

        gothic.Store = store

        goth.UseProviders(
            google.New(googleClientID, googleClientSecret, callbackUrl),
        )
    })

    return &AuthHandler{
        GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
        GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
        CallbackUrl:        os.Getenv("CALLBACK_URL"),
        JwtSecret:          os.Getenv("JWT_SECRET"),
        Store:              store,
    }
}