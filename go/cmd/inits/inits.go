package inits

import (
	"fmt"
    "os"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// technically only using supabase for DB here so could just use postgrest
// but this kinda just makes sense like we're using SUPABASE not just postgrest
// and we might use other supabase features in the future so just use supabase client
// but do note this is community maintained client; either way you can access any 
// supabase feature via rest api so this is just convenience wrapper really
func InitializePostgresClient() (*pgxpool.Pool, error) {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println("Unable to connect to database:", err)
		return nil, err
	}

	if err := dbpool.Ping(ctx); err != nil {
		fmt.Println("Unable to ping database:", err)
		return nil, err
	}

	fmt.Println("Connected to database at", os.Getenv("DATABASE_URL"))

	return dbpool, nil
}

func InitializeRedisClient() (*redis.Client, error) {
	ctx := context.Background()

	// later move this to full URL auth that embeds username and password
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
		Password: "", // no password set
		DB: 0,  // use default DB
	})
	
	err := rdb.Ping(ctx).Err()
	if err != nil {
		fmt.Println("Unable to connect to redis:", err)
		return nil, err
	}

	fmt.Println("Connected to redis at", os.Getenv("REDIS_URL"))

	return rdb, nil
}