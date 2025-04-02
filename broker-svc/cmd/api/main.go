package main

import (
	"fmt"
	"log"
	"net/http"
)

const webPort = "8888"

// Config struct to hold app configuration and methods
type Config struct{}

func main() {
	// Initialize the app configuration
	app := &Config{}

	// Set up the server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	log.Println("Broker service started on port: ", webPort)

	// Run the server
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}
