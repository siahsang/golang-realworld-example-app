package main

import (
	"context"
	"database/sql"
	"github.com/siahsang/blog/internal/data"
	"log"
	"time"
)

type application struct {
	models data.Models
}

func main() {

	db, err := openDBConnection()

	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("Error closing database connection: %v", err)
		}
	}()

	app := application{
		models: data.NewModels(),
	}
}

func openDBConnection() (*sql.DB, error) {
	db, err := sql.Open("postgres", "")
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(10)

	duration, err := time.ParseDuration("10s")
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	context.WithTimeout(context.Background(), 5*time.Second)
	err = db.PingContext(context.Background())
	if err != nil {
		return nil, err
	}

	return db, nil

}
