package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/example/memory-architecture-sample/internal/application/chat"
	httpapi "github.com/example/memory-architecture-sample/internal/interfaces/http"
	"github.com/example/memory-architecture-sample/internal/infrastructure/memory"
	"github.com/example/memory-architecture-sample/internal/infrastructure/responder"
)

func main() {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	memoryStore := memory.NewInMemoryStore()
	chatService := chat.NewService(memoryStore, responder.NewTemplateResponder(), time.Now)
	server := httpapi.NewServer(chatService)

	log.Printf("API listening on http://localhost:%s", port)
	log.Printf("Swagger UI: http://localhost:%s/swagger", port)
	if err := http.ListenAndServe(":"+port, server.Handler()); err != nil {
		log.Fatal(err)
	}
}
