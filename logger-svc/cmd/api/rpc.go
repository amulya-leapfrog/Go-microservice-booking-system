package main

import (
	"context"
	"log"
	"time"
)

type RPCServer struct{}

type RPCPayload struct {
	Name string
	Data string
}

type LogEntry struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name" json:"name"`
	Data      string    `bson:"data" json:"data"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func (r *RPCServer) LogInfoViaRPC(payload RPCPayload, resp *string) error {
	collection := client.Database("logs").Collection("logs")

	_, err := collection.InsertOne(context.TODO(), LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Println("Error inserting into logs via RPC: ", err)
		return err
	}

	*resp = "Successfully logged via RPC: " + payload.Name + " - " + payload.Data
	return nil
}
