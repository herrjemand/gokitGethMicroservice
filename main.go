package main

import (
    "log"
    "time"
    "net/http"
    "github.com/herrjemand/gethGoKitRPCMicroService/router"
    "github.com/herrjemand/gethGoKitRPCMicroService/proto"

    "net"
    "context"
    "google.golang.org/grpc"
)

const serverAddress string = "127.0.0.1"
const httpServerPort string = "8080"
const grpcServerPort string = "9090"

func main() {
    svc := router.EthServiceImp{}

    errors := make(chan error)
    go func() {
        router := router.GenerateHTTPRouter(svc)
        server := &http.Server{
            Handler:      router.(http.Handler),
            Addr:         serverAddress + ":" + httpServerPort,

            WriteTimeout: 15 * time.Second,
            ReadTimeout:  15 * time.Second,
        }
        log.Println("Starting HTTP server at " + serverAddress + ":" + httpServerPort + "...")

        errors <- server.ListenAndServe()
    }()

    ctx := context.Background()

    go func() {
        listener, err := net.Listen("tcp", serverAddress + ":" + grpcServerPort)
        if err != nil {
            errors <- err
            return
        }

        gRPCServer := grpc.NewServer()
        proto.RegisterEthGRPCServer(gRPCServer, router.GetGethGRPCEndpoints(ctx, svc))

        log.Println("Starting gRPC server at " + serverAddress + ":" + grpcServerPort + "...")

        errors <- gRPCServer.Serve(listener)
    }()

    log.Fatal(<-errors)
}
