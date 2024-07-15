package main

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"

	pbAuth "github.com/Far-sa/microservices-repo/common/genproto/common/proto/auth"
	pbAuthz "github.com/Far-sa/microservices-repo/common/genproto/common/proto/authz"
	pbUser "github.com/Far-sa/microservices-repo/common/genproto/common/proto/user"
	// !_ "github.com/Far-sa/common/docs/openapi/common/proto"
)

func main() {
	// gRPC Gateway setup
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				UseEnumNumbers:  false,
				AllowPartial:    false,
				EmitUnpopulated: true,
				Indent:          "  ",
			},
		}),
	)

	opts := []grpc.DialOption{grpc.WithInsecure()}

	// Register Auth Service
	if err := pbAuth.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts); err != nil {
		log.Fatalf("Failed to register auth service: %v", err)
	}

	// Register User Service
	if err := pbUser.RegisterUserServiceHandlerFromEndpoint(ctx, mux, "localhost:50052", opts); err != nil {
		log.Fatalf("Failed to register user service: %v", err)
	}

	// Register Authz Service
	if err := pbAuthz.RegisterAuthzServiceHandlerFromEndpoint(ctx, mux, "localhost:50053", opts); err != nil {
		log.Fatalf("Failed to register authz service: %v", err)
	}

	// Swagger UI setup
	statikFS, err := fs.New()
	if err != nil {
		log.Fatalf("Failed to create statik filesystem: %v", err)
	}

	swaggerMux := http.NewServeMux()
	swaggerMux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(statikFS)))

	// Start gRPC Gateway
	go func() {
		log.Println("Starting gRPC Gateway on port 4000")
		if err := http.ListenAndServe(":4000", mux); err != nil {
			log.Fatalf("Failed to serve gRPC Gateway: %v", err)
		}
	}()

	// Start Swagger UI server
	log.Println("Serving Swagger UI at http://localhost:8080/swagger/")
	if err := http.ListenAndServe(":8080", swaggerMux); err != nil {
		log.Fatalf("Failed to serve Swagger UI: %v", err)
	}

}
