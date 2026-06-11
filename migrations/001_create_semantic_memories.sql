CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS semantic_memories (
    id text PRIMARY KEY,
    conversation_id text NOT NULL,
    content text NOT NULL,
    embedding vector(32) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS semantic_memories_conversation_idx
    ON semantic_memories (conversation_id);

CREATE INDEX IF NOT EXISTS semantic_memories_embedding_idx
    ON semantic_memories USING hnsw (embedding vector_cosine_ops);
