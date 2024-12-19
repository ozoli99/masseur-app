package main

import (
	"log"
	"net/http"

	"github.com/ozoli99/Kaida/api"
	"github.com/ozoli99/Kaida/db"
	"github.com/ozoli99/Kaida/service"
)

func main() {
    var database db.Database = &db.SQLiteDatabase{}
    if err := database.InitializeDatabase(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    appointmentService := service.AppointmentService{Database: database}

    webSocketServer := api.NewWebSocketServer()
    api.StartWebSocketServer(webSocketServer, "8081")

    httpServer := api.Server{
        AppointmentService: &appointmentService,
        WebSocketServer: webSocketServer,
    }

    httpServer.AddMiddleware(api.LoggingMiddleware)
    httpServer.AddMiddleware(api.CORSMiddleware)

    http.Handle("/", http.FileServer(http.Dir("./static")))

    log.Println("Starting HTTP server on :8080...")
    if err := httpServer.StartServer("8080"); err != nil {
        log.Fatalf("Failed to start HTTP server: %v", err)
    }
}