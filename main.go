package main

import (
    "log"
    "time"
    "net/http"
    // "google.golang.org/grpc"
    "github.com/herrjemand/gethGoKitRPCMicroService/router"
)

const httpServerAddress string = "127.0.0.1"
const httpServerPort string = "8000"

func main() {
    router := router.GenerateRouter()
    server := &http.Server{
        Handler:      router.(http.Handler),
        Addr:         httpServerAddress + ":" + httpServerPort,

        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    log.Println("Starting HTTP server at " + httpServerAddress + ":" + httpServerPort + "...")
    log.Fatal(server.ListenAndServe())
}
