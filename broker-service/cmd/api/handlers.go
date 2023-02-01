package main

import (
	"broker-service/auths"
	"broker-service/event"
	"broker-service/logs"
	"broker-service/mails"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type MailPayload struct {
	Name    string `json:"name"`
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type AuthPayload struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type RabbitPayload struct {
	Severity string
	Data     any
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

// HandleSubmission is the main point of entry into the broker. It accepts a JSON
// payload and performs an action based on the value of "action" in that JSON.
func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth-json":
		app.authenticate(w, requestPayload.Auth)
	case "auth-rabbit":
		app.rabbitRequest(w, requestPayload.Auth, "auth.CHECK", "Authenticated via RabbitMQ")
	case "auth-rpc":
		app.rpcRequest(w, "authentication-service:5001", "RPCServer.AuthenticateViaRPC", requestPayload.Auth)
	case "auth-grpc":
		app.authenticateViaGRPC(w, requestPayload.Auth)
	case "log-json":
		app.logItem(w, requestPayload.Log)
	case "log-rabbit":
		app.rabbitRequest(w, requestPayload.Log, "log.INFO", "Logged via RabbitMQ")
	case "log-rpc":
		app.rpcRequest(w, "logger-service:5001", "RPCServer.LogInfo", requestPayload.Log)
	case "log-grpc":
		app.logViaGRPC(w, requestPayload.Log)
	case "mail-json":
		app.sendMail(w, requestPayload.Mail)
	case "mail-rabbit":
		app.rabbitRequest(w, requestPayload.Mail, "mail.SEND", "Mail sent via RabbitMQ")
	case "mail-rpc":
		app.rpcRequest(w, "mailer-service:5001", "RPCServer.SendMailViaRPC", requestPayload.Mail)
	case "mail-grpc":
		app.sendMailViaGRPC(w, requestPayload.Mail)
	default:
		app.errorJSON(w, errors.New("unknown action"))
	}
}

// logItem logs an event to the database
func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged via JSON"

	app.writeJSON(w, http.StatusAccepted, payload)
}

// authenticate calls the authentication microservice and sends back the appropriate response
func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// call the service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code
	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling auth service"))
		return
	}

	// create a variable we'll read response.Body into
	var jsonFromService jsonResponse

	// decode the json from the auth service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated via JSON"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) sendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData, _ := json.MarshalIndent(msg, "", "\t")

	// call the mail service
	mailServiceURL := "http://mailer-service/send"

	// post to mail service
	request, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the right status code
	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling mail service"))
		return
	}

	// send back json
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Message sent via JSON to " + msg.To

	app.writeJSON(w, http.StatusAccepted, payload)
}

// rabbitRequest pushes the data to RabbitMQ.
func (app *Config) rabbitRequest(w http.ResponseWriter, rabbitPayload any, severity, msg string) {
	rpaylod := RabbitPayload{
		Severity: severity,
		Data:     rabbitPayload,
	}

	err := app.pushToQueue(rpaylod)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = msg

	app.writeJSON(w, http.StatusAccepted, payload)
}

// pushToQueue pushes a message into RabbitMQ
func (app *Config) pushToQueue(payload RabbitPayload) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), payload.Severity)
	if err != nil {
		return err
	}

	return nil
}

func (app *Config) rpcRequest(w http.ResponseWriter, addr string, serviceMethod string, rpcPayload any) {
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var result []byte
	err = client.Call(serviceMethod, rpcPayload, &result)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	_ = json.Unmarshal(result, &payload)

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logViaGRPC(w http.ResponseWriter, requestPayload LogPayload) {
	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Name,
			Data: requestPayload.Data,
		},
	})
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged via gRPC"

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) sendMailViaGRPC(w http.ResponseWriter, requestPayload MailPayload) {
	conn, err := grpc.Dial("mailer-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

	c := mails.NewMailServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.SendMail(ctx, &mails.MailRequest{
		MailEntry: &mails.Mail{
			From:    requestPayload.From,
			To:      requestPayload.To,
			Subject: requestPayload.Subject,
			Message: requestPayload.Message,
		},
	})
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Mail sent via gRPC"

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) authenticateViaGRPC(w http.ResponseWriter, requestPayload AuthPayload) {
	conn, err := grpc.Dial("authentication-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

	c := auths.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.Authenticate(ctx, &auths.AuthRequest{
		AuthEntry: &auths.Auth{
			Name:    requestPayload.Name,
			Email:      requestPayload.Email,
			Password: requestPayload.Password,
		},
	})
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated via gRPC"

	app.writeJSON(w, http.StatusAccepted, payload)
}