package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/juhun32/jtracker-backend/utils"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"github.com/markbates/goth/gothic"
	amqp "github.com/rabbitmq/amqp091-go"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// simple struct for returning user data to frontend or other handlers
type User struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Handler struct {
	AuthHandler     *utils.AuthHandler
	firestoreClient *firestore.Client
	rabbitCh        *amqp.Channel
	rabbitQ         amqp.Queue
}

// initialize a new handler with an AuthHandler (implementation in utils/main.go) and Firestore client
// authHandler parameter passed in from cmd/main.go
//
//	reason: gorilla/mux spins up a new goroutine for each request
//	        so, we pass in the same AuthHandler to each handler to ensure global state is maintained
func NewHandler(
	firestoreClient *firestore.Client,
	authHandler *utils.AuthHandler,
	rabbitCh *amqp.Channel,
	rabbitQ amqp.Queue,
) *Handler {
	return &Handler{
		AuthHandler:     authHandler,
		firestoreClient: firestoreClient,
		rabbitCh:        rabbitCh,
		rabbitQ:         rabbitQ,
	}
}

// {provider} is a variable that can be anything (if we want more providers in the future)
// in this case, we only support google
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/auth/{provider}", h.Auth).Methods("GET").Name("auth")
	router.HandleFunc("/auth/{provider}/callback", h.AuthProviderCallback).Methods("GET").Name("authProviderCallback")
	router.HandleFunc("/auth/{provider}/logout", h.Logout).Methods("GET").Name("logout")
}

// note: gothic uses a global Store variable so we can just directly call gothic
func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] Auth [*]")
	log.Println("-----------------")
	provider := mux.Vars(r)["provider"]
	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

	// if the user is already authenticated, redirect them to their dashboard
	user, err := IsAuthenticated(r)
	if err == nil {
		fmt.Println("user already authenticated", user)
		http.Redirect(w, r, "http://localhost:5173/dashboard", http.StatusFound)
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

	// for some reason CompleteUserAuth isn't properly setting session values so we have to do it manually
	// this is to guarantee that the user receives a session
	session, _ := h.AuthHandler.Store.Get(r, "session")
	session.Values["user_id"] = user.UserID
	session.Values["email"] = user.Email
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// check if user exists in Firestore
	userExists, err := CheckUserExists(user.Email, h.firestoreClient, r.Context())
	if err != nil {
		fmt.Println("Error checking if user exists:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !userExists {
		// add user to firestore (gmail document id)
		// by default, firestore will create a new document if it doesnt exist
		_, err = h.firestoreClient.Collection("users").Doc(user.Email).Set(r.Context(), map[string]interface{}{
			"email": user.Email,
		})
		if err != nil {
			fmt.Printf("Error adding user to Firestore: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defaultApplication := map[string]interface{}{
			"company":     "Google",
			"role":        "Software Engineer Intern",
			"location":    "Mountain View, CA",
			"appliedDate": "2021-01-01",
			"status":      "Applied",
			"link":        "https://www.google.com",
		}

		// Add default application to user/{email}/applications
		doc, _, err := h.firestoreClient.Collection("users").Doc(user.Email).Collection("applications").Add(r.Context(), defaultApplication)
		if err != nil {
			fmt.Printf("Error adding default application to Firestore: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send the default application to RabbitMQ
		message := map[string]interface{}{
			"email":       user.Email,
			"objectID":    doc.ID,
			"company":     "Google",
			"role":        "Software Engineer Intern",
			"location":    "Mountain View, CA",
			"appliedDate": "2021-01-01",
			"status":      "Applied",
			"link":        "https://www.google.com",
			"operation":   "add", // operation to perform in algolia
		}

		messageBody, err := json.Marshal(message)
		if err != nil {
			fmt.Printf("Error marshalling message: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Publish message to RabbitMQ
		err = utils.PublishWithRetry(h.rabbitCh, "", h.rabbitQ.Name, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        messageBody,
		})
		if err != nil {
			fmt.Printf("Error publishing message after retries: %v\n", err)
		} else {
			log.Println("Message published")
			log.Println("-----------------")
		}
	}

	http.Redirect(w, r, "http://localhost:5173/dashboard", http.StatusFound)
}

func CheckUserExists(userEmail string, firestoreClient *firestore.Client, ctx context.Context) (bool, error) {
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

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	log.Println("[*] Logout [*]")
	log.Println("-----------------")

	provider := mux.Vars(r)["provider"]
	if provider != "google" {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

	// same as AuthProviderCallback, for some reason gothic's logout function isn't properly clearing session
	// so let's just do it manually. atp why wouldnt i just not use a library lmfao
	session, err := h.AuthHandler.Store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = -1
	session.Values = make(map[interface{}]interface{})

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "http://localhost:5173", http.StatusTemporaryRedirect)
}

// use the request and gothic.Store.Get to see if user is authed
// once again, something's wrong with gothic session handling so we have to
// manually get the session from Store and check if email exists
// it would be nice to implement this as a middleware but for now just call it directly within each route
func IsAuthenticated(r *http.Request) (*User, error) {
	log.Println("[*] IsAuthenticated [*]")
	log.Println("-----------------")

	session, err := gothic.Store.Get(r, "session")
	if err != nil {
		return nil, err
	}

	emailValue, ok := session.Values["email"]
	if !ok {
		return nil, fmt.Errorf("email not found in session")
	}

	email, ok := emailValue.(string)
	if !ok {
		return nil, fmt.Errorf("email is not a string")
	}

	log.Println("Authenticated")
	log.Println("-----------------")

	return &User{Email: email}, nil
}
