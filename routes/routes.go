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

	// Subrouter for /api routes
	apiRouter := r.PathPrefix("/api").Subrouter()

	// Public routes under /api
	apiRouter.HandleFunc("/login", api.LoginHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/forms", api.FormHandler).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/register", api.RegisterHandler).Methods("POST", "OPTIONS")

	// Initialize database connection
	db.InitDB()

	return r
}
