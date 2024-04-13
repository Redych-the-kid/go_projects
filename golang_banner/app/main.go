package main

import (
	"context"
	"database/sql"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"os"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", databaseURL)
	
	if err != nil {
		panic(err)
	}
	
	defer db.Close()
	
	cache := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",
		DB:       0,
	})
	ctx := context.Background()
	var e = echo.New()
	server := &Server{
		tokens: map[string]string{
			"IGOTTHEPOWER!": "admin",
			"IMACREEP":      "user",
		},
		db:    db,
		cache: cache,
		ctx:   ctx,
	}
	
	RegisterHandlers(e, server)
	e.Start(":8080")
}
