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

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"github.com/markbates/goth/gothic"
	"github.com/golang-jwt/jwt/v5"
	
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	AuthHandler     *utils.AuthHandler
	firestoreClient *firestore.Client
}

// initialize a new handler with an AuthHandler (implementation in utils/main.go) and Firestore client
// authHandler parameter passed in from cmd/main.go
//
//	reason: gorilla/mux spins up a new goroutine for each request
//	        so, we pass in the same AuthHandler to each handler to ensure global state is maintained
func NewHandler(
	firestoreClient *firestore.Client,
	authHandler *utils.AuthHandler,
) *Handler {
	return &Handler{
		AuthHandler:     authHandler,
		firestoreClient: firestoreClient,
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

	// if the user is already authenticated, redirect them to their dashboard
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	user, err := IsAuthenticated(r)
	if err == nil {
		fmt.Println("user already authenticated", user)
		http.Redirect(w, r, frontendURL + "/dashboard", http.StatusFound)
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

	// at this point, user is verified to be authed and we have
	// 1. made a session (locally with gothic)
	// 2. made a JWT (to send to frontend)

	// check if user exists in Firestore
	userExists, err := checkUserExists(user.Email, h.firestoreClient, r.Context())
	if err != nil {
		fmt.Println("Error checking if user exists:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !userExists {
		// add user to firestore (gmail document id)
		// by default, firestore will create a new document if it doesnt exist
		// no need to create a default application subcollection since it will be created on first add application request
		_, err = h.firestoreClient.Collection("users").Doc(user.Email).Set(r.Context(), map[string]interface{}{
			"email":             user.Email,
			"applicationsCount": 0,
		})
		if err != nil {
			fmt.Printf("Error adding user to Firestore: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	http.Redirect(w, r, frontendURL + "/auth-complete?token=" + tokenString, http.StatusFound)
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

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	
	http.Redirect(w, r, frontendURL + "/logout-complete", http.StatusTemporaryRedirect)
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

func checkUserExists(userEmail string, firestoreClient *firestore.Client, ctx context.Context) (bool, error) {
	_, err := firestoreClient.Collection("users").Doc(userEmail).Get(ctx) // queries users collection for document username

	if err != nil {
		if status.Code(err) == codes.NotFound { // not a real error
			return false, nil
		} else { // any other error
			return false, err
		}
	} else { // no error == user found
		return true, nil
	}
}