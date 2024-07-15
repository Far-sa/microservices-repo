package main

import (
	"context"
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"log"
	"net"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"

	pbAuth "github.com/Far-sa/microservices-repo/common/genproto/common/proto/auth"
	pbUser "github.com/Far-sa/microservices-repo/common/genproto/common/proto/user"
)

type userServiceServer struct {
	publicKey *rsa.PublicKey
	pbUser.UnimplementedUserServiceServer
}

func (s *userServiceServer) RegisterUser(ctx context.Context, req *pbUser.RegisterUserRequest) (*pbUser.RegisterUserResponse, error) {
	// Register user (for simplicity, assume success)
	return &pbUser.RegisterUserResponse{Success: true}, nil
}

func (s *userServiceServer) GetUserProfile(ctx context.Context, req *pbUser.GetUserProfileRequest) (*pbUser.UserProfileResponse, error) {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return s.publicKey, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	username := claims["username"].(string)
	// Fetch user profile from database (for simplicity, return static data)
	return &pbUser.UserProfileResponse{
		Username: username,
		Email:    "user@example.com",
	}, nil
}

func main() {
	// Fetch the public key from AuthService
	conn, err := grpc.Dial("auth-service:50051", grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	if err != nil {
		log.Fatalf("Failed to connect to AuthService: %v", err)
	}
	defer conn.Close()

	authService := pbAuth.NewAuthServiceClient(conn)
	resp, err := authService.GetPublicKey(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Fatalf("Failed to get public key: %v", err)
	}

	block, _ := pem.Decode([]byte(resp.PublicKey))
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(block.Bytes)
	if err != nil {
		log.Fatalf("Failed to parse public key: %v", err)
	}

	grpcServer := grpc.NewServer()
	pbUser.RegisterUserServiceServer(grpcServer, &userServiceServer{
		publicKey: publicKey,
	})

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen on port 50052: %v", err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
