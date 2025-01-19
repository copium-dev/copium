package user

import (
	"fmt"
	"net/http"
	"encoding/json"

	"github.com/juhun32/jtracker-backend/service/auth"

	"github.com/gorilla/mux"
	"cloud.google.com/go/firestore"
)

type Handler struct {
	firestoreClient *firestore.Client
}

func NewHandler(firestoreClient *firestore.Client) *Handler {
	return &Handler{
		firestoreClient: firestoreClient,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/user/dashboard", h.Dashboard).Methods("GET").Name("dashboard")
}

// current implementation is TEMPORARY!!!!
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
    user, err := auth.IsAuthenticated(r)

    if err != nil {
        fmt.Printf("Error: %v\n", err)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    fmt.Println("user", user)

	// the actual implementation of this will use the user object from auth.IsAuthenticated
	// and query Firestore for the user's data and return to client 

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

