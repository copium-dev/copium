package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/copium-dev/copium/go/utils"

	"github.com/gorilla/mux"
	"github.com/markbates/goth/gothic"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	authHandler *utils.AuthHandler
	pgClient *pgxpool.Pool
	frontendURL string
}

func NewHandler(authHandler *utils.AuthHandler, pgClient *pgxpool.Pool, frontendURL string) *Handler {
	return &Handler{
		authHandler: authHandler,
		pgClient: pgClient,
		frontendURL: frontendURL,
	}
}

// {provider} is a variable that can be anything (if we want more providers in the future)
// in this case, we only support google
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/auth/{provider}", h.Auth).Methods("GET").Name("auth")
	router.HandleFunc("/auth/{provider}/callback", h.AuthProviderCallback).Methods("GET").Name("authProviderCallback")
	router.HandleFunc("/auth/{provider}/logout", h.Logout).Methods("GET").Name("logout")
}

// gothic is JUST to handle oauth flow, since cross-domain cookies are a pain to deal with
// technically, we could just set up a custom domain but Cloud Run custom domains are in preview mode, so not ideal
func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] Auth [*]")
	log.Println("-----------------")
	provider := mux.Vars(r)["provider"]
	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

	user, err := IsAuthenticated(r)
	if err == nil {
		fmt.Println("user already authenticated", user)
		http.Redirect(w, r, h.frontendURL + "/dashboard", http.StatusFound)
		return
	}

	gothic.BeginAuthHandler(w, r)

	log.Println("Auth complete")
	log.Println("-----------------")
}

func (h *Handler) AuthProviderCallback(w http.ResponseWriter, r *http.Request) {
	provider := mux.Vars(r)["provider"]
	if provider != "google" {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Printf("Auth error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// this sucks but in prod we can't send cookies across domains, and Cloud Run custom domains
	// are only in preview mode, so we have to make and sign a JWT and send to frontend
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 24 * 30).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// rpc call to postgres (on supabase) to create user and update login time
	// note: QueryRow is used when we expect single row results
	var success bool
	err = h.pgClient.QueryRow(r.Context(), "SELECT service.login($1)", user.Email).Scan(&success)
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Failed to create or update user", http.StatusInternalServerError)
		return
	}

	if !success {
		log.Println("Failed to create or update user")
		http.Error(w, "Failed to create or update user", http.StatusInternalServerError)
		return
	}

	log.Println("User created or updated successfully")
	log.Println("-----------------")

	http.Redirect(w, r, h.frontendURL + "/auth-complete?token=" + tokenString, http.StatusFound)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] Logout [*]")
	log.Println("-----------------")

	provider := mux.Vars(r)["provider"]
	if provider != "google" {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
	
	http.Redirect(w, r, h.frontendURL + "/logout-complete", http.StatusTemporaryRedirect)
}

// check for authentication using JWT
// the key change here vs. the original is that we don't use gothic for auth verification or session management
// since we create our own JWTs. so, gothic is JUST to handle the oauth flow
func IsAuthenticated(r *http.Request) (string, error) {
	log.Println("[*] IsAuthenticated [*]")
	log.Println("-----------------")

    // get token from Authorization header
    authHeader := r.Header.Get("Authorization")
    if !strings.HasPrefix(authHeader, "Bearer ") {
        return "", fmt.Errorf("no token provided")
    }
    
    // extract token value
    tokenString := strings.TrimPrefix(authHeader, "Bearer ")
    
    // parse and validate token
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(os.Getenv("JWT_SECRET")), nil
    })
    
    if err != nil {
        return "", fmt.Errorf("invalid token: %v", err)
    }
    
	// checks if token was signed w/ secret key and not tampered
	// also checks if not expired
    if !token.Valid {
        return "", fmt.Errorf("token is not valid")
    }
    
	// get claims so we can extract email
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return "", fmt.Errorf("invalid token claims")
    }
    
    email, ok := claims["email"].(string)
    if !ok || email == "" {
        return "", fmt.Errorf("email not found in token")
    }
    
    log.Println("Authenticated via JWT")
    log.Println("-----------------")
    
    return email, nil
}