package responder

import (
	"context"
	"fmt"

	"github.com/example/memory-architecture-sample/internal/domain"
)

type TemplateResponder struct{}

func NewTemplateResponder() TemplateResponder {
	return TemplateResponder{}
}

func (TemplateResponder) Reply(_ context.Context, userMessage string, recent []domain.Message) (string, error) {
	if len(recent) == 0 {
		return fmt.Sprintf("This is our first remembered message. You said: %q", userMessage), nil
	}
	return fmt.Sprintf(
		"I remember %d recent messages in this conversation. You said: %q",
		len(recent),
		userMessage,
	), nil
}
