package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/aman1218/portal_bapp/config"
	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() {
	cfg := config.AppConfig.Database
	dataSourceName := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	var err error
	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database connected!")
}

func CloseDB() {
	db.Close()
}
