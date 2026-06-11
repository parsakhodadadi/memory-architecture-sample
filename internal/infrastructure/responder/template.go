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

func (TemplateResponder) Reply(
	_ context.Context,
	userMessage string,
	recent []domain.Message,
	recalled []domain.SemanticMemory,
) (string, error) {
	if len(recent) == 0 {
		if len(recalled) == 0 {
			return fmt.Sprintf("This is our first remembered message. You said: %q", userMessage), nil
		}
		return fmt.Sprintf("I recalled %q. You said: %q", recalled[0].Content, userMessage), nil
	}
	response := fmt.Sprintf(
		"I remember %d recent messages in this conversation. You said: %q",
		len(recent),
		userMessage,
	)
	if len(recalled) > 0 {
		response += fmt.Sprintf(" A related long-term memory is: %q.", recalled[0].Content)
	}
	return response, nil
}
