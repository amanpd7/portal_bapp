package main

import (
	"log"
	"net/http"

	"github.com/aman1218/portal_bapp/config"
	"github.com/aman1218/portal_bapp/db"
	"github.com/aman1218/portal_bapp/routes"
)

func main() {
	log.Println("Loading config...")
	config.LoadConfig()

	log.Println("Initializing database connection...")
	db.InitDB()
	defer db.CloseDB()

	r := routes.Router()

	port := config.AppConfig.Server.Port
	log.Printf("Starting server on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
