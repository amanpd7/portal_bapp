package routes

import (
	"github.com/aman1218/portal_bapp/api"
	"github.com/aman1218/portal_bapp/db"
	"github.com/aman1218/portal_bapp/middleware"
	"github.com/gorilla/mux"
)

func Router() *mux.Router {

	r := mux.NewRouter()

	// Middleware
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.CORSMiddleware)

	// Public routes
	r.HandleFunc("/login", api.LoginHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/forms", api.FormHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/register", api.RegisterHandler).Methods("POST", "OPTIONS")

	// Initialize database connection
	db.InitDB()
	return r
}
