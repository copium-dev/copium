package api

import (
    "log"
    "net/http"

    "github.com/gorilla/mux"
    "github.com/rs/cors"
    "github.com/juhun32/jtracker-backend/service/user"
    "github.com/juhun32/jtracker-backend/service/auth"
    "github.com/juhun32/jtracker-backend/utils"
    "cloud.google.com/go/firestore"
)

type APIServer struct {
    addr            string
    firestoreClient *firestore.Client
    authHandler     *utils.AuthHandler
}

func NewAPIServer(addr string, firestoreClient *firestore.Client, authHandler *utils.AuthHandler) *APIServer {
    return &APIServer{
        addr:            addr,
        firestoreClient: firestoreClient,
        authHandler:     authHandler,
    }
}

// initialize router, database, and other dependencies
func (s *APIServer) Run() error {
    router := mux.NewRouter()

    log.Println("Listening on", s.addr)

    // only user data requires firestore client
    userHandler := user.NewHandler(s.firestoreClient)
    userHandler.RegisterRoutes(router)

    authHandler := auth.NewHandler(s.firestoreClient, s.authHandler)
    authHandler.RegisterRoutes(router)

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