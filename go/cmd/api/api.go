package api

import (
    "log"
    "net/http"
	"os"
	
	"github.com/copium-dev/copium/go/service/user"
    "github.com/copium-dev/copium/go/service/auth"
	"github.com/copium-dev/copium/go/service/postings"
    "github.com/copium-dev/copium/go/utils"
    
	"github.com/gorilla/mux"
    "github.com/rs/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type APIServer struct {
	addr string
	pgClient *pgxpool.Pool
	redisClient *redis.Client
	authHandler *utils.AuthHandler
}

func NewAPIServer(
	addr string, pgClient *pgxpool.Pool, redisClient *redis.Client, authHandler *utils.AuthHandler,
) *APIServer {
    return &APIServer{
        addr: addr,
		pgClient: pgClient,
		redisClient: redisClient,
		authHandler: authHandler,
    }
}

// initialize router, database, and other dependencies
func (s *APIServer) Run() error {
    router := mux.NewRouter()

    log.Println("Listening on", s.addr)

    userHandler := user.NewHandler(s.pgClient, s.redisClient)
    userHandler.RegisterRoutes(router)

	// authHandler requires frontendURL
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

    authHandler := auth.NewHandler(s.authHandler, s.pgClient, frontendURL)
    authHandler.RegisterRoutes(router)

	// postings handler requires nada but use same pattern for consistency & future-proofing for future features
	postingsHandler := postings.NewHandler()
	postingsHandler.RegisterRoutes(router)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://www.copium.dev", "https://copium.dev", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true, 
		MaxAge:           86400,
	})

    // wrap router with the CORS handler
    handler := c.Handler(router)

    return http.ListenAndServe(s.addr, handler)
}