package main

import (
	"database/sql"
	"fmt"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "homuhomu"
	dbname   = "postgres"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var e = echo.New()
	server := &Server{
		tokens: map[string]string{
			"IGOTTHEPOWER!": "admin",
			"IMACREEP":      "user",
		},
		db: db,
	}
	RegisterHandlers(e, server)
	e.Start(":8080")
}
