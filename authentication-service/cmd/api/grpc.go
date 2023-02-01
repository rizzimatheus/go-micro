package main

import (
	"authentication-service/auths"
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

type AuthServer struct {
	auths.UnimplementedAuthServiceServer
}

func (a *AuthServer) Authenticate(ctx context.Context, req *auths.AuthRequest) (*auths.AuthResponse, error) {
	input := req.GetAuthEntry()

	// validate the user against the database
	user, err := app.Models.User.GetByEmail(input.Email)
	if err != nil {
		res := &auths.AuthResponse{Result: "invalid credentials"}
		return res, err
	}

	valid, err := user.PasswordMatches(input.Password)
	if err != nil || !valid {
		res := &auths.AuthResponse{Result: "invalid credentials"}
		return res, err
	}

	// log authentication
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		res := &auths.AuthResponse{Result: "faild log request"}
		return res, err
	}

	// return response
	res := &auths.AuthResponse{Result: "Authenticated via gRPC"}
	return res, nil
}

func (app *Config) gRPCListen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", gRpcPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	s := grpc.NewServer()

	auths.RegisterAuthServiceServer(s, &AuthServer{})

	log.Printf("gRPC Server started on port %s", gRpcPort)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}
}
