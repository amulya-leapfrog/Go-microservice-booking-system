package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// routes method to configure the app's routes
func (app *Config) routes() http.Handler {
	// Create a new router instance
	mux := chi.NewRouter()

	// Apply CORS middleware
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"}, // Allow all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-Id"},
		AllowCredentials: true,
		ExposedHeaders:   []string{"Link"},
		MaxAge:           300, // Cache the preflight response for 5 minutes
	}))

	// Apply heartbeat middleware for health check
	mux.Use(middleware.Heartbeat("/health"))

	// Define routes and handlers
	mux.Get("/", app.Broker)
	mux.Post("/handle", app.HandleSubmission)

	return mux
}
