package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

var mySigningKey = []byte(os.Getenv("TOKEN_SECRET"))

type RequestPayload struct {
	Action      string             `json:"action"`
	Auth        AuthRequest        `json:"auth,omitempty"`
	Reservation ReservationRequest `json:"reservation,omitempty"`
}

type AuthRequest struct {
	Action   string      `json:"action"`
	AuthData AuthPayload `json:"authData"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	FullName string `json:"fullName,omitempty"`
	Password string `json:"password"`
}

type ReservationRequest struct {
	Action          string          `json:"action"`
	ReservationData ReservationData `json:"reservationData"`
}

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

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "reserve":
		app.reservation(w, r, requestPayload.Reservation)
	default:
		app.errorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthRequest) {
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	request, err := http.NewRequest("POST", "http://auth-svc:8181/auth", bytes.NewBuffer(jsonData))

	if err != nil {
		log.Printf("Error creating auth service request: %v\n", err)
		app.errorJSON(w, fmt.Errorf("error calling authentication service"))
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Error calling auth service: %v\n", err)
		app.errorJSON(w, fmt.Errorf("error calling authentication service"))
		return
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusOK {
		app.errorJSON(w, errors.New("error calling auth service"))
		return
	}

	var auhResponse authResponse

	err = json.NewDecoder(response.Body).Decode(&auhResponse)
	if err != nil {
		log.Printf("Error reading auth service response: %v\n", err)
		app.errorJSON(w, err)
		return
	}

	if auhResponse.Error {
		app.errorJSON(w, errors.New(auhResponse.Message))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = fmt.Sprintf("Authentication Service!: %s", auhResponse.Message)

	if a.Action == "signup" {
		app.writeJSON(w, http.StatusAccepted, payload)
		return
	}

	token, err := generateToken(auhResponse.Data.ID)
	if err != nil {
		log.Printf("Error generating token: %v\n", err)
		app.errorJSON(w, err)
		return
	}

	payload.Data = token

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) reservation(w http.ResponseWriter, r *http.Request, reservationReq ReservationRequest) {
	tokenString, err := extractToken(r)
	if err != nil {
		log.Printf("Error extracting token: %v\n", err)
		app.errorJSON(w, fmt.Errorf("invalid token"))
		return
	}

	id, err := verifyJWT(tokenString)
	if err != nil {
		log.Printf("Error verifying JWT: %v\n", err)
		app.errorJSON(w, fmt.Errorf("unauthorized"))
		return
	}

	reservationReq.ReservationData.UserId = id

	switch reservationReq.Action {
	case "add":
		app.createReservation(w, reservationReq.ReservationData)
	default:
		app.errorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) createReservation(w http.ResponseWriter, rd ReservationData) {
	client, err := rpc.Dial("tcp", "reservation-svc:5002")
	if err != nil {
		log.Println("Error connecting to reservation rpc from broker: ", err)
	}

	var rpcPayload RPCPayload
	rpcPayload.ReservationData.UserId = rd.UserId
	rpcPayload.ReservationData.RestaurantID = rd.RestaurantID
	rpcPayload.ReservationData.Count = rd.Count
	rpcPayload.ReservationData.ReservationTime = rd.ReservationTime
	rpcPayload.ReservationData.Remarks = rd.Remarks

	var result string
	err = client.Call("RPCServer.CreateReservation", rpcPayload, &result)
	if err != nil {
		log.Println("Error sending payload to reservation rpc from broker: ", err)
		app.errorJSON(w, fmt.Errorf("error creating reservation booking"))
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = fmt.Sprintf("Reservation Service!: %s", result)
	app.writeJSON(w, http.StatusOK, payload)
}

// ExtractToken extracts and returns the JWT token from the Authorization header
func extractToken(r *http.Request) (string, error) {
	// Fetch the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header missing")
	}

	// Split the "Bearer <token>" into two parts
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return "", fmt.Errorf("authorization header format is invalid")
	}

	// Return the token part
	return tokenParts[1], nil
}

// generateToken generates a JWT token
func generateToken(id string) (string, error) {
	// Create the claims
	claims := jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(time.Hour * 1).Unix(), // Expiration time
		"iat": time.Now().Unix(),                    // Issued At
	}

	// Create a new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyJWT verifies the token and returns the claims (fullName and email in this case)
func verifyJWT(tokenString string) (string, error) {
	// Parse the token and validate it with the signing key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token method is what we expect (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return mySigningKey, nil
	})
	if err != nil {
		return "", err
	}

	// Return the token if it's valid
	if token.Valid {
		// Get the claims (assuming they are stored in MapClaims)
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return "", fmt.Errorf("invalid claims")
		}

		// Retrieve user id from the claims
		id, idOk := claims["id"].(string)

		if !idOk {
			return "", fmt.Errorf("userId not found in token")
		}

		// Return the userID
		return id, nil
	}
	return "", fmt.Errorf("invalid token")
}
