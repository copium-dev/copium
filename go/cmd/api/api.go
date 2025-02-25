package api

import (
    "log"
    "net/http"
    
	"github.com/juhun32/jtracker-backend/service/user"
    "github.com/juhun32/jtracker-backend/service/auth"
	"github.com/juhun32/jtracker-backend/service/postings"
    "github.com/juhun32/jtracker-backend/utils"
    
	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
    "github.com/rs/cors"
	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"cloud.google.com/go/pubsub"
)

type APIServer struct {
    addr string
    firestoreClient *firestore.Client
	algoliaClient *search.APIClient
    authHandler *utils.AuthHandler
	pubsubClient *pubsub.Client
}

func NewAPIServer(addr string,
	firestoreClient *firestore.Client,
	algoliaClient *search.APIClient,
	authHandler *utils.AuthHandler,
	pubsubClient *pubsub.Client,
) *APIServer {
    return &APIServer{
        addr: addr,
        firestoreClient: firestoreClient,
		algoliaClient: algoliaClient,
        authHandler: authHandler,
		pubsubClient: pubsubClient,
    }
}

// initialize router, database, and other dependencies
func (s *APIServer) Run() error {
    router := mux.NewRouter()

    log.Println("Listening on", s.addr)

    userHandler := user.NewHandler(s.firestoreClient, s.algoliaClient, s.pubsubClient)
    userHandler.RegisterRoutes(router)

    authHandler := auth.NewHandler(s.firestoreClient, s.authHandler)
    authHandler.RegisterRoutes(router)

	postingsHandler := postings.NewHandler(s.algoliaClient)
	postingsHandler.RegisterRoutes(router)

    // create new CORS handler
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:5173"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"*"},
        AllowCredentials: true,
    })

    // wrap router with the CORS handler
    handler := c.Handler(router)

    return http.ListenAndServe(s.addr, handler)
}