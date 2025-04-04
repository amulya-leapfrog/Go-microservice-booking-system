package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

const rpcPort = "5002"

var maxAcceptError int

var conn *sql.DB

func main() {
	maxAcceptErrorStr := os.Getenv("MAX_ACCEPT_ERROR")

	var err error
	maxAcceptError, err = strconv.Atoi(maxAcceptErrorStr)
	if err != nil {
		log.Fatalf("Error converting MAX_ACCEPT_ERROR to integer: %v", err)
	}

	conn = connectToPostgres()
	if conn == nil {
		log.Fatal("Can't connect to Postgres")
	}

	defer conn.Close()

	err = rpc.Register(new(RPCServer))
	if err != nil {
		log.Panic("Error registering Rservation RPC server: ", err)
	}

	if err := rpcListen(); err != nil {
		log.Panic("Rservation RPC server exited with error: ", err)
	}
}

func rpcListen() error {
	log.Println("Starting Reservation RPC server on port: ", rpcPort)
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		log.Println("Error starting RPC server: ", err)
		return err
	}
	defer listen.Close()

	acceptFailures := 0

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			acceptFailures++

			if acceptFailures >= maxAcceptError {
				log.Printf("Too many accept failures (%d). Shutting down Reservation RPC server.\n", acceptFailures)
				return fmt.Errorf("exceeded max accept failures")
			}

			// small sleep to avoid spinning too fast
			time.Sleep(200 * time.Millisecond)
			continue
		}

		// reset error count on successful accept
		acceptFailures = 0
		go rpc.ServeConn(conn)
	}
}

func connectToPostgres() *sql.DB {
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
