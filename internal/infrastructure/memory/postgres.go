package memory

import (
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/memory-architecture-sample/internal/domain"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	store := &PostgresStore{pool: pool}
	if err := store.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if err := store.ensureSchema(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

func (s *PostgresStore) Save(
	ctx context.Context,
	memory domain.SemanticMemory,
	embedding []float32,
) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO semantic_memories
			(id, conversation_id, content, embedding, created_at, expires_at)
		VALUES ($1, $2, $3, $4::vector, $5, $6)`,
		memory.ID,
		memory.ConversationID,
		memory.Content,
		vectorLiteral(embedding),
		memory.CreatedAt,
		memory.ExpiresAt,
	)
	return err
}

func (s *PostgresStore) Search(
	ctx context.Context,
	conversationID string,
	query string,
	embedding []float32,
	limit int,
) ([]domain.SemanticMemory, error) {
	candidateLimit := limit * 20
	if candidateLimit < 100 {
		candidateLimit = 100
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, conversation_id, content, created_at, expires_at,
		       1 - (embedding <=> $2::vector) AS similarity
		FROM semantic_memories
		WHERE conversation_id = $1 AND expires_at > now()
		ORDER BY embedding <=> $2::vector
		LIMIT $3`,
		conversationID,
		vectorLiteral(embedding),
		candidateLimit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memories := make([]domain.SemanticMemory, 0)
	for rows.Next() {
		var memory domain.SemanticMemory
		if err := rows.Scan(
			&memory.ID,
			&memory.ConversationID,
			&memory.Content,
			&memory.CreatedAt,
			&memory.ExpiresAt,
			&memory.Similarity,
		); err != nil {
			return nil, err
		}
		memories = append(memories, memory)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rankMemories(query, memories, limit), nil
}

func (s *PostgresStore) Delete(ctx context.Context, memoryID string) (bool, error) {
	result, err := s.pool.Exec(ctx, `DELETE FROM semantic_memories WHERE id = $1`, memoryID)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

func (s *PostgresStore) ClearConversation(ctx context.Context, conversationID string) error {
	_, err := s.pool.Exec(
		ctx,
		`DELETE FROM semantic_memories WHERE conversation_id = $1`,
		conversationID,
	)
	return err
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) ensureSchema(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
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
	`)
	return err
}

func vectorLiteral(values []float32) string {
	var builder strings.Builder
	builder.WriteByte('[')
	for index, value := range values {
		if index > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(strconv.FormatFloat(float64(value), 'f', 6, 32))
	}
	builder.WriteByte(']')
	return builder.String()
}
