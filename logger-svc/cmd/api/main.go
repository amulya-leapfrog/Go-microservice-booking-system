package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// port constants
const (
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
)

var client *mongo.Client

var maxAcceptError int

func main() {
	maxAcceptErrorStr := os.Getenv("MAX_ACCEPT_ERROR")

	var err error
	maxAcceptError, err = strconv.Atoi(maxAcceptErrorStr)
	if err != nil {
		log.Fatalf("Error converting MAX_ACCEPT_ERROR to integer: %v", err)
	}

	mongoClient, err := connectToMongo()
	if err != nil {
		log.Fatal("Error connecting to MongoDB: ", err)
	}
	client = mongoClient

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal("Error disconnecting from MongoDB: ", err)
		}
	}()

	err = rpc.Register(new(RPCServer))
	if err != nil {
		log.Panic("Error registering Logger RPC server: ", err)
	}

	if err := rpcListen(); err != nil {
		log.Panic("Logger RPC server exited with error: ", err)
	}
}

func rpcListen() error {
	log.Println("Starting Logger RPC server on port: ", rpcPort)
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
				log.Printf("Too many accept failures (%d). Shutting down Logger RPC server.\n", acceptFailures)
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

func connectToMongo() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoURL)

	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Error connecting to mongo: ", err)
		return nil, err
	}

	log.Println("Successfully connected to the mongo!")

	return c, err
}
