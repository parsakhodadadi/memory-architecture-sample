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
internal/config/         environment configuration
internal/httpapi/        routes, handlers, and Swagger
internal/memory/         memory contracts and implementations
internal/chat/           chat orchestration
internal/retention/      cleanup job
migrations/              PostgreSQL and pgvector schema
docs/                    architecture and saved chat notes
```
