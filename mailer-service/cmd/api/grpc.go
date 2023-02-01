package main

import (
	"context"
	"fmt"
	"log"
	"mailer-service/mails"
	"net"

	"google.golang.org/grpc"
)

type MailServer struct {
	mails.UnimplementedMailServiceServer
}

func (m *MailServer) SendMail(ctx context.Context, req *mails.MailRequest) (*mails.MailResponse, error) {
	input := req.GetMailEntry()

	msg := Message{
		From:    input.From,
		To:      input.To,
		Subject: input.Subject,
		Data:    input.Message,
	}

	err := app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		res := &mails.MailResponse{Result: "failed"}
		return res, err
	}

	// return response
	res := &mails.MailResponse{Result: "Mail sent via gRPC"}
	return res, nil
}

func (app *Config) gRPCListen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", gRpcPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	s := grpc.NewServer()

	mails.RegisterMailServiceServer(s, &MailServer{})

	log.Printf("gRPC Server started on port %s", gRpcPort)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}
}