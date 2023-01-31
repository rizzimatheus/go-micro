package main

import (
	"encoding/json"
	"log"
)

// RPCServer is the type for our RPC Server. Methods that take this as a receiver are available
// over RPC, as long as they are exported.
type RPCServer struct{}

// RPCPayload is the type for data we receive from RPC
type RPCPayload struct {
	Name    string
	From    string
	To      string
	Subject string
	Message string
}

type RPCResponse struct {
	Error   bool
	Message string
	Data    any
}

// LogInfo writes our payload to mongo
func (r *RPCServer) SendMailViaRPC(payload RPCPayload, resp *[]byte) error {
	msg := Message{
		From:    payload.From,
		To:      payload.To,
		Subject: payload.Subject,
		Data:    payload.Message,
	}

	err := app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		log.Println("error sending mail", err)
		return err
	}

	// resp is the message sent back to the RPC caller
	*resp, _ = json.Marshal(RPCResponse{
		Error:   false,
		Message: "Mail sent via RPC to " + payload.To,
	})

	return nil
}
