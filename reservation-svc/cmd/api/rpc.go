package main

import (
	"context"
	"fmt"
	"log"
	"net/rpc"
	"time"
)

type RPCServer struct{}

type ReservationData struct {
	ReservationID   string    `json:"id,omitempty"`
	RestaurantID    string    `json:"restaurantID,omitempty"`
	UserId          string    `json:"userID,omitempty"`
	Count           string    `json:"count,omitempty"`
	ReservationTime string    `json:"reservationTime,omitempty"`
	Remarks         string    `json:"remarks,omitempty"`
	CreatedAt       time.Time `json:"createdAt,omitempty"`
}

type RPCPayload struct {
	ReservationData ReservationData
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

const dbTimeout = time.Second * 3

func (r *RPCServer) CreateReservation(payload RPCPayload, resp *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var newID string

	stmt := `INSERT INTO reservations (restaurant_id, user_id, count, reservation_time, remarks)
	VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := conn.QueryRowContext(ctx, stmt,
		payload.ReservationData.RestaurantID,
		payload.ReservationData.UserId,
		payload.ReservationData.Count,
		payload.ReservationData.ReservationTime,
		payload.ReservationData.Remarks,
	).Scan(&newID)
	if err != nil {
		log.Println("Error inserting into reservations via RPC: ", err)
		return err
	}

	successMsg := fmt.Sprintf("Reservation: %s successfully created for userID: %s", newID, payload.ReservationData.UserId)

	log.Println(successMsg)

	var logPayload LogPayload
	logPayload.Name = "Reservation_Created"
	logPayload.Data = successMsg

	logItemViaRPC(logPayload)

	*resp = "Reservation created successfully"
	return nil
}

func logItemViaRPC(l LogPayload) {
	client, err := rpc.Dial("tcp", "logger-svc:5001")
	if err != nil {
		log.Println("Error connecting to logger rpc from reservation: ", err)
	}

	var result string
	err = client.Call("RPCServer.LogInfoViaRPC", l, &result)
	if err != nil {
		log.Println("Error sending payload to logger rpc from reservation: ", err)
	}

	log.Println(result)
}
