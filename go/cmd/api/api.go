package api

import (
    "log"
    "net/http"
	"os"
    
	"github.com/copium-dev/copium/go/service/user"
    "github.com/copium-dev/copium/go/service/auth"
	"github.com/copium-dev/copium/go/service/postings"
    "github.com/copium-dev/copium/go/utils"
    
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
	pubsubTopic *pubsub.Topic
	orderingKey string
}

func NewAPIServer(addr string,
	firestoreClient *firestore.Client,
	algoliaClient *search.APIClient,
	authHandler *utils.AuthHandler,
	pubsubTopic *pubsub.Topic,
	orderingKey string,
) *APIServer {
    return &APIServer{
        addr: addr,
        firestoreClient: firestoreClient,
		algoliaClient: algoliaClient,
        authHandler: authHandler,
		pubsubTopic: pubsubTopic,
		orderingKey: orderingKey,
    }
}

// initialize router, database, and other dependencies
func (s *APIServer) Run() error {
    router := mux.NewRouter()

    log.Println("Listening on", s.addr)

    userHandler := user.NewHandler(s.firestoreClient, s.algoliaClient, s.pubsubTopic, s.orderingKey)
    userHandler.RegisterRoutes(router)

    authHandler := auth.NewHandler(s.firestoreClient, s.authHandler)
    authHandler.RegisterRoutes(router)

	postingsHandler := postings.NewHandler(s.algoliaClient)
	postingsHandler.RegisterRoutes(router)

    // create new CORS handler
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{os.Getenv("FRONTEND_URL"), "http://localhost:5173"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"*"},
        AllowCredentials: true,
    })

    // wrap router with the CORS handler
    handler := c.Handler(router)

    return http.ListenAndServe(s.addr, handler)
}