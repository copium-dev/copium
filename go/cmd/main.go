package main

import (
    "log"
    "os"

    "github.com/copium-dev/copium/go/cmd/api"
	"github.com/copium-dev/copium/go/cmd/inits"
	"github.com/copium-dev/copium/go/utils"
)

func main() {
	pgClient, err := inits.InitializePostgresClient()
	if err != nil {
		log.Fatal(err)
	}

	redisClient, err := inits.InitializeRedisClient()
	if err != nil {
		log.Fatal(err)
	}

	// cloud run will provide PORT=8080 by default in env
    port := os.Getenv("PORT")
    if port == "" {
        port = "8000"
    }

    log.Printf("Starting server on port %s", port)

	server := api.NewAPIServer(":" + port, pgClient, redisClient, utils.NewAuthHandler())
    if err := server.Run(); err != nil {
        log.Fatal(err)
    }
}