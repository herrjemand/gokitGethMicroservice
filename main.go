package main

import (
    "log"
    "time"
    "net/http"
    "github.com/herrjemand/gethGoKitRPCMicroService/router"

    // "net"
    // "google.golang.org/grpc"
)

const httpServerAddress string = "127.0.0.1"
const httpServerPort string = "8080"
const grpcServerPort string = "9090"

func main() {
    errors := make(chan error)
    go func() {
        router := router.GenerateHTTPRouter()
        server := &http.Server{
            Handler:      router.(http.Handler),
            Addr:         httpServerAddress + ":" + httpServerPort,

            WriteTimeout: 15 * time.Second,
            ReadTimeout:  15 * time.Second,
        }
        log.Println("Starting HTTP server at " + httpServerAddress + ":" + httpServerPort + "...")

        errors <- server.ListenAndServe()
    }()


    log.Fatal(<-errors)
}
