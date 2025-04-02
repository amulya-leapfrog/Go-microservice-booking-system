package main

import (
	"authentication/data"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
)

type RequestPayload struct {
	Action   string      `json:"action"`
	AuthData AuthPayload `json:"authData"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	FullName string `json:"fullName,omitempty"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var responsePayload jsonResponse

	if requestPayload.Action == "login" {
		var newUser *data.User

		// check for user in database
		newUser, err := newUser.GetByEmail(requestPayload.AuthData.Email)
		if err != nil {
			log.Printf("Error getting user for login: %v\n", err)
			app.errorJSON(w, err, http.StatusUnauthorized)
			return
		}

		// password match
		_, err = newUser.PasswordMatches(requestPayload.AuthData.Password)
		if err != nil {
			log.Printf("Login error: %v for id: %d\n", err, newUser.ID)
			app.errorJSON(w, fmt.Errorf("invalid credentials"), http.StatusUnauthorized)
			return
		}

		loginMsg := fmt.Sprintf("User with id: %d logged in", newUser.ID)
		log.Print(loginMsg)

		// log user login
		logData := LogPayload{
			Name: "Auth_Login",
			Data: loginMsg,
		}

		app.logItemViaRPC(logData)

		// return login success
		responsePayload.Data.Email = newUser.Email
		responsePayload.Data.FullName = newUser.FullName
		responsePayload.Message = "Login success"
		app.writeJSON(w, http.StatusOK, responsePayload)
	} else {
		// check for user in database
		var newUser *data.User

		// check for user in database
		newUser, err := newUser.GetByEmail(requestPayload.AuthData.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("Error getting user: %v\n", err)
			app.errorJSON(w, err)
			return
		}

		if newUser != nil {
			log.Printf("User with id: %d already exists\n", newUser.ID)
			responsePayload.Error = true
			responsePayload.Message = "User already exists"
			app.writeJSON(w, http.StatusOK, responsePayload)
			return
		}

		// create user
		id, err := newUser.Insert(data.User{
			Email:    requestPayload.AuthData.Email,
			Password: requestPayload.AuthData.Password,
			FullName: requestPayload.AuthData.FullName,
		})
		if err != nil {
			log.Printf("Error inserting user: %v\n", err)
			responsePayload.Error = true
			responsePayload.Message = "Error inserting user"
			app.writeJSON(w, http.StatusAccepted, responsePayload)
			return
		}

		signupMsg := fmt.Sprintf("New User with id: %d created", id)
		log.Print(signupMsg)

		// log user signup via RPC
		logData := LogPayload{
			Name: "Auth_Signup",
			Data: signupMsg,
		}

		app.logItemViaRPC(logData)

		// return signup success
		responsePayload.Message = "Signup success"
		app.writeJSON(w, http.StatusOK, responsePayload)
	}

}

func (app *Config) logItemViaRPC(l LogPayload) {
	client, err := rpc.Dial("tcp", "logger-svc:5001")
	if err != nil {
		log.Println("Error connecting to logger rpc from auth: ", err)
	}

	var rpcPayload RPCPayload
	rpcPayload.Name = l.Name
	rpcPayload.Data = l.Data

	var result string
	err = client.Call("RPCServer.LogInfoViaRPC", rpcPayload, &result)
	if err != nil {
		log.Println("Error sending payload to logger rpc from auth: ", err)
	}

	log.Println(result)
}
