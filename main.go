package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gd-harco/chirpy/internal/api"
	"github.com/gd-harco/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	cfg := api.NewConfig(database.New(db), os.Getenv("PLATFORM"), os.Getenv("SECRET_KEY"))

	serv := http.Server{
		Addr:    ":8080",
		Handler: api.NewMux(cfg),
	}
	_ = serv.ListenAndServe()
}
