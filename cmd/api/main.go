package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/example/memory-architecture-sample/internal/application/chat"
	memoryapp "github.com/example/memory-architecture-sample/internal/application/memory"
	"github.com/example/memory-architecture-sample/internal/config"
	"github.com/example/memory-architecture-sample/internal/infrastructure/embedding"
	httpapi "github.com/example/memory-architecture-sample/internal/interfaces/http"
	"github.com/example/memory-architecture-sample/internal/infrastructure/memory"
	"github.com/example/memory-architecture-sample/internal/infrastructure/responder"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	shortTerm := memory.NewRedisStore(
		cfg.RedisAddress,
		cfg.RedisPassword,
		cfg.ShortTermTTL,
		cfg.ShortTermLimit,
	)
	defer shortTerm.Close()
	if err := shortTerm.Ping(ctx); err != nil {
		log.Fatalf("connect to Redis: %v", err)
	}

	longTerm, err := memory.NewPostgresStore(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("connect to PostgreSQL: %v", err)
	}
	defer longTerm.Close()

	embedder := embedding.NewHashEmbedder()
	chatService := chat.NewService(
		shortTerm,
		longTerm,
		embedder,
		responder.NewTemplateResponder(),
		time.Now,
		cfg.LongTermTTL,
	)
	memoryService := memoryapp.NewService(longTerm, embedder)
	server := httpapi.NewServer(chatService, memoryService)

	log.Printf("API listening on http://localhost:%s", cfg.HTTPPort)
	log.Printf("Swagger UI: http://localhost:%s/swagger", cfg.HTTPPort)
	if err := http.ListenAndServe(":"+cfg.HTTPPort, server.Handler()); err != nil {
		log.Fatal(err)
	}
}
