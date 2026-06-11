package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/memory-architecture-sample/internal/application/chat"
	httpapi "github.com/example/memory-architecture-sample/internal/interfaces/http"
	"github.com/example/memory-architecture-sample/internal/infrastructure/memory"
	"github.com/example/memory-architecture-sample/internal/infrastructure/responder"
)

func newTestHandler() http.Handler {
	service := chat.NewService(
		memory.NewInMemoryStore(),
		responder.NewTemplateResponder(),
		time.Now,
	)
	return httpapi.NewServer(service).Handler()
}

func TestChatAndHistoryEndpoints(t *testing.T) {
	handler := newTestHandler()
	body := []byte(`{"conversationId":"demo","message":"Remember that I use Go."}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}

	historyRequest := httptest.NewRequest(http.MethodGet, "/api/v1/conversations/demo/messages", nil)
	historyResponse := httptest.NewRecorder()
	handler.ServeHTTP(historyResponse, historyRequest)
	if historyResponse.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", historyResponse.Code)
	}

	var result struct {
		Messages []json.RawMessage `json:"messages"`
	}
	if err := json.Unmarshal(historyResponse.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if len(result.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result.Messages))
	}
}

func TestSwaggerAndOpenAPIEndpoints(t *testing.T) {
	handler := newTestHandler()
	for _, path := range []string{"/swagger", "/openapi.json"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s returned %d", path, response.Code)
		}
		if response.Body.Len() == 0 {
			t.Fatalf("%s returned an empty body", path)
		}
	}
}
