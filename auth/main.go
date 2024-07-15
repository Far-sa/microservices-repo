package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net"
	"time"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pbAuth "github.com/Far-sa/microservices-repo/common/genproto/common/proto/auth"
)

type authServiceServer struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	pbAuth.UnimplementedAuthServiceServer
}

func generateKeyPair(bits int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func marshalPublicKey(publicKey *rsa.PublicKey) (string, error) {
	pubASN1, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})
	return string(pubBytes), nil
}

func (s *authServiceServer) Login(ctx context.Context, req *pbAuth.LoginRequest) (*pbAuth.LoginResponse, error) {
	// Validate user credentials (for simplicity, assume success)
	claims := jwt.MapClaims{
		"username": req.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return nil, err
	}
	return &pbAuth.LoginResponse{Token: tokenString}, nil
}

func (s *authServiceServer) GetPublicKey(ctx context.Context, req *emptypb.Empty) (*pbAuth.PublicKeyResponse, error) {
	// Return the public key in PEM format
	pubKeyPEM, err := marshalPublicKey(s.publicKey)
	if err != nil {
		return nil, err
	}
	return &pbAuth.PublicKeyResponse{PublicKey: string(pubKeyPEM)}, nil
}

func main() {
	privateKey, err := generateKeyPair(2048)
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}
	publicKey := &privateKey.PublicKey

	grpcServer := grpc.NewServer()
	pbAuth.RegisterAuthServiceServer(grpcServer, &authServiceServer{
		privateKey: privateKey,
		publicKey:  publicKey,
	})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen on port 50051: %v", err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
