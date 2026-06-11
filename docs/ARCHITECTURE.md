# Architecture

## Goal

Demonstrate how a chatbot can remember recent messages, retrieve relevant older
facts, and automatically forget expired data.

## Components

```text
Swagger / HTTP client
        |
        v
     Go API
        |
        +-- Chat service
        |     +-- Short-term memory -> Redis
        |     +-- Long-term memory  -> PostgreSQL + pgvector
        |     +-- Local response generator
        |
        +-- Retention service
              +-- Redis TTL
              +-- PostgreSQL expires_at cleanup
```

## Memory Flow

1. The API receives a chat message.
2. A deterministic embedding is generated locally.
3. Redis returns recent messages for the conversation.
4. pgvector returns semantically similar long-term memories.
5. The local response generator creates a testable reply from that context.
6. The user message and reply are stored in Redis.
7. Important user messages can also be stored in PostgreSQL.

## Retention

- Short-term messages expire automatically through Redis TTL.
- Long-term memories have an `expires_at` timestamp.
- A background cleanup job deletes expired PostgreSQL rows.

## Project Layout

```text
cmd/api/                 application entry point
internal/domain/         enterprise entities
internal/ports/          interfaces owned by the application core
internal/application/    use cases and orchestration
internal/infrastructure/ technical adapter implementations
internal/interfaces/http HTTP handlers and Swagger delivery
migrations/              PostgreSQL and pgvector schema
docs/                    architecture and saved chat notes
```

Dependencies point inward: HTTP and storage adapters depend on application
ports and domain entities. The application layer does not import HTTP, Redis,
PostgreSQL, or any other infrastructure package.
