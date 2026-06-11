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
7. The user message and its 32-dimensional embedding are stored in PostgreSQL.

## Memory Modules

- `ShortTermMemory` is implemented by Redis using one bounded list per
  conversation. Every write refreshes the list TTL.
- `LongTermMemory` is implemented by PostgreSQL. pgvector orders memories by
  cosine distance to the query embedding.
- `HashEmbedder` is deterministic and local. It is intentionally simple: this
  project demonstrates memory architecture rather than model quality.
- Search uses hybrid ranking: pgvector retrieves candidates, then normalized
  word overlap reranks them. Punctuation, common question words, and simple
  aliases such as `like`, `prefer`, and `favorite` are normalized locally.

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
