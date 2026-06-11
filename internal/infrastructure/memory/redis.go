package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/example/memory-architecture-sample/internal/domain"
)

type RedisStore struct {
	client *redis.Client
	ttl    time.Duration
	limit  int64
}

func NewRedisStore(address, password string, ttl time.Duration, limit int64) *RedisStore {
	return &RedisStore{
		client: redis.NewClient(&redis.Options{
			Addr:     address,
			Password: password,
		}),
		ttl:   ttl,
		limit: limit,
	}
}

func (s *RedisStore) Save(ctx context.Context, messages ...domain.Message) error {
	if len(messages) == 0 {
		return nil
	}

	key := redisKey(messages[0].ConversationID)
	values := make([]any, 0, len(messages))
	for _, message := range messages {
		data, err := json.Marshal(message)
		if err != nil {
			return err
		}
		values = append(values, data)
	}

	pipe := s.client.TxPipeline()
	pipe.RPush(ctx, key, values...)
	pipe.LTrim(ctx, key, -s.limit, -1)
	pipe.Expire(ctx, key, s.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisStore) List(ctx context.Context, conversationID string, limit int) ([]domain.Message, error) {
	start := int64(-limit)
	if limit <= 0 {
		start = -s.limit
	}
	values, err := s.client.LRange(ctx, redisKey(conversationID), start, -1).Result()
	if err != nil {
		return nil, err
	}

	messages := make([]domain.Message, 0, len(values))
	for _, value := range values {
		var message domain.Message
		if err := json.Unmarshal([]byte(value), &message); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func (s *RedisStore) Clear(ctx context.Context, conversationID string) error {
	return s.client.Del(ctx, redisKey(conversationID)).Err()
}

func (s *RedisStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}

func redisKey(conversationID string) string {
	return fmt.Sprintf("chat:working-memory:%s", conversationID)
}
