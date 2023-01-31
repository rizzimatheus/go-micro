package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// RPCServer is the type for our RPC Server. Methods that take this as a receiver are available
// over RPC, as long as they are exported.
type RPCServer struct{}

// RPCPayload is the type for data we receive from RPC
type RPCPayload struct {
	Name     string
	Email    string
	Password string
}

type RPCResponse struct {
	Error   bool
	Message string
	Data    any
}

// LogInfo writes our payload to mongo
func (r *RPCServer) AuthenticateViaRPC(payload RPCPayload, resp *[]byte) error {
	// validate the user against the database
	user, err := app.Models.User.GetByEmail(payload.Email)
	if err != nil {
		log.Println("invalid credentials", err)
		return err
	}

	valid, err := user.PasswordMatches(payload.Password)
	if err != nil || !valid {
		log.Println("invalid credentials", err)
		return err
	}

	// log authentication
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		log.Println("error logging authentication", err)
		return err
	}

	// resp is the message sent back to the RPC caller
	*resp, _ = json.Marshal(RPCResponse{
		Error:   false,
		Message: "Authenticated via RPC in user " + user.Email,
		Data:    user,
	})

	return nil
}
