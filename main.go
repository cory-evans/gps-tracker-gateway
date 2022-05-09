package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	auth "go.buf.build/jonwhitty/go-grpc-gateway/corux/gps-tracker-auth/auth/v1"
	position "go.buf.build/jonwhitty/go-grpc-gateway/corux/gps-tracker-position/position/v1"
)

var (
	listenPort                 = os.Getenv("LISTEN_PORT")
	authGrpcServerEndpoint     = os.Getenv("AUTH_SERVICE")
	positionGrpcServerEndpoint = os.Getenv("POSITION_SERVICE")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	authMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := auth.RegisterAuthServiceHandlerFromEndpoint(ctx, authMux, authGrpcServerEndpoint, opts)
	if err != nil {
		return err
	}

	positionMux := runtime.NewServeMux()
	err = position.RegisterPositionServiceHandlerFromEndpoint(ctx, positionMux, positionGrpcServerEndpoint, opts)
	if err != nil {
		return err
	}

	serverMux := http.NewServeMux()
	serverMux.Handle("/auth/", http.StripPrefix("/auth", authMux))
	serverMux.Handle("/position/", http.StripPrefix("/position", positionMux))

	serverMux.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("./public"))))

	srv := &http.Server{
		Addr:    ":" + listenPort,
		Handler: cors(serverMux),
	}

	return srv.ListenAndServe()
}

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	})
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
