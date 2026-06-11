package chat

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/example/memory-architecture-sample/internal/domain"
	"github.com/example/memory-architecture-sample/internal/ports"
)

var (
	ErrConversationIDRequired = errors.New("conversationId is required")
	ErrMessageRequired        = errors.New("message is required")
)

type Service struct {
	memory    ports.ShortTermMemory
	responder ports.Responder
	now       func() time.Time
}

type SendInput struct {
	ConversationID string
	Message        string
}

type SendOutput struct {
	ConversationID string           `json:"conversationId"`
	Reply          string           `json:"reply"`
	Context        []domain.Message `json:"context"`
}

func NewService(memory ports.ShortTermMemory, responder ports.Responder, now func() time.Time) *Service {
	return &Service{memory: memory, responder: responder, now: now}
}

func (s *Service) Send(ctx context.Context, input SendInput) (SendOutput, error) {
	conversationID := strings.TrimSpace(input.ConversationID)
	if conversationID == "" {
		return SendOutput{}, ErrConversationIDRequired
	}

	message := strings.TrimSpace(input.Message)
	if message == "" {
		return SendOutput{}, ErrMessageRequired
	}

	recent, err := s.memory.List(ctx, conversationID, 10)
	if err != nil {
		return SendOutput{}, err
	}

	reply, err := s.responder.Reply(ctx, message, recent)
	if err != nil {
		return SendOutput{}, err
	}

	createdAt := s.now().UTC()
	userMessage := domain.Message{
		ID:             newID(),
		ConversationID: conversationID,
		Role:           domain.RoleUser,
		Content:        message,
		CreatedAt:      createdAt,
	}
	assistantMessage := domain.Message{
		ID:             newID(),
		ConversationID: conversationID,
		Role:           domain.RoleAssistant,
		Content:        reply,
		CreatedAt:      createdAt,
	}

	if err := s.memory.Save(ctx, userMessage, assistantMessage); err != nil {
		return SendOutput{}, err
	}

	return SendOutput{
		ConversationID: conversationID,
		Reply:          reply,
		Context:        recent,
	}, nil
}

func (s *Service) History(ctx context.Context, conversationID string, limit int) ([]domain.Message, error) {
	if strings.TrimSpace(conversationID) == "" {
		return nil, ErrConversationIDRequired
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.memory.List(ctx, conversationID, limit)
}

func (s *Service) Clear(ctx context.Context, conversationID string) error {
	if strings.TrimSpace(conversationID) == "" {
		return ErrConversationIDRequired
	}
	return s.memory.Clear(ctx, conversationID)
}

func newID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(value[:])
}
