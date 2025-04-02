package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

const webPort = "8181"

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {

	conn := connectToDB()
	if conn == nil {
		log.Panicln("Can't connect to Postgres")
	}

	defer conn.Close()

	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	log.Println("Authentication service started on port: ", webPort)

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func connectToDB() *sql.DB {
	connStr := os.Getenv("DSN")

	// Open the database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	// Check if the connection is valid
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database: ", err)
	}

	log.Println("Successfully connected to the postgres!")

	return db
}
