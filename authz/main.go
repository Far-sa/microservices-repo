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
	pbAuthz "github.com/Far-sa/microservices-repo/common/genproto/common/proto/authz"
)

type authzServiceServer struct {
	publicKey *rsa.PublicKey
	pbAuthz.UnimplementedAuthzServiceServer
}

func (s *authzServiceServer) CheckPermission(ctx context.Context, req *pbAuthz.CheckPermissionRequest) (*pbAuthz.CheckPermissionResponse, error) {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return s.publicKey, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// claims
	_, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if the user has permission for the action (for simplicity, allow all actions)
	allowed := true

	return &pbAuthz.CheckPermissionResponse{Allowed: allowed}, nil
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
	pbAuthz.RegisterAuthzServiceServer(grpcServer, &authzServiceServer{
		publicKey: publicKey,
	})

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Failed to listen on port 50053: %v", err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
