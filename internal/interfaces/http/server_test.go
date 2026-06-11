package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/memory-architecture-sample/internal/application/chat"
	memoryapp "github.com/example/memory-architecture-sample/internal/application/memory"
	"github.com/example/memory-architecture-sample/internal/infrastructure/embedding"
	httpapi "github.com/example/memory-architecture-sample/internal/interfaces/http"
	"github.com/example/memory-architecture-sample/internal/infrastructure/memory"
	"github.com/example/memory-architecture-sample/internal/infrastructure/responder"
)

func newTestHandler() http.Handler {
	shortTerm := memory.NewInMemoryStore()
	longTerm := memory.NewInMemoryLongTermStore()
	embedder := embedding.NewHashEmbedder()
	service := chat.NewService(
		shortTerm,
		longTerm,
		embedder,
		responder.NewTemplateResponder(),
		time.Now,
		30*24*time.Hour,
	)
	memoryService := memoryapp.NewService(longTerm, embedder)
	return httpapi.NewServer(service, memoryService).Handler()
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

func TestSemanticMemorySearchEndpoint(t *testing.T) {
	handler := newTestHandler()
	chatBody := []byte(`{"conversationId":"demo","message":"My favorite database is PostgreSQL."}`)
	chatRequest := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(chatBody))
	chatRequest.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(httptest.NewRecorder(), chatRequest)

	searchBody := []byte(`{"conversationId":"demo","query":"Which database do I like?","limit":5}`)
	searchRequest := httptest.NewRequest(http.MethodPost, "/api/v1/memories/search", bytes.NewReader(searchBody))
	searchRequest.Header.Set("Content-Type", "application/json")
	searchResponse := httptest.NewRecorder()
	handler.ServeHTTP(searchResponse, searchRequest)

	if searchResponse.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", searchResponse.Code, searchResponse.Body.String())
	}
	var result struct {
		Memories []json.RawMessage `json:"memories"`
	}
	if err := json.Unmarshal(searchResponse.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if len(result.Memories) == 0 {
		t.Fatal("expected at least one semantic memory")
	}
}

func TestDeleteOneSemanticMemoryEndpoint(t *testing.T) {
	handler := newTestHandler()
	chatBody := []byte(`{"conversationId":"demo","message":"My favorite database is PostgreSQL."}`)
	chatRequest := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(chatBody))
	chatRequest.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(httptest.NewRecorder(), chatRequest)

	searchBody := []byte(`{"conversationId":"demo","query":"database","limit":5}`)
	searchRequest := httptest.NewRequest(http.MethodPost, "/api/v1/memories/search", bytes.NewReader(searchBody))
	searchRequest.Header.Set("Content-Type", "application/json")
	searchResponse := httptest.NewRecorder()
	handler.ServeHTTP(searchResponse, searchRequest)

	var searchResult struct {
		Memories []struct {
			ID string `json:"id"`
		} `json:"memories"`
	}
	if err := json.Unmarshal(searchResponse.Body.Bytes(), &searchResult); err != nil {
		t.Fatalf("invalid search response: %v", err)
	}
	if len(searchResult.Memories) != 1 {
		t.Fatalf("expected one memory, got %d", len(searchResult.Memories))
	}

	deleteRequest := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/memories/"+searchResult.Memories[0].ID,
		nil,
	)
	deleteResponse := httptest.NewRecorder()
	handler.ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", deleteResponse.Code, deleteResponse.Body.String())
	}
}

func TestClearConversationSemanticMemoryEndpoint(t *testing.T) {
	handler := newTestHandler()
	for _, conversationID := range []string{"first", "second"} {
		body := []byte(`{"conversationId":"` + conversationID + `","message":"Remember this value."}`)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(httptest.NewRecorder(), request)
	}

	clearRequest := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/conversations/first/memories",
		nil,
	)
	clearResponse := httptest.NewRecorder()
	handler.ServeHTTP(clearResponse, clearRequest)
	if clearResponse.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", clearResponse.Code)
	}

	for conversationID, expected := range map[string]int{"first": 0, "second": 1} {
		body := []byte(`{"conversationId":"` + conversationID + `","query":"value","limit":5}`)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/memories/search", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		var result struct {
			Memories []json.RawMessage `json:"memories"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatalf("invalid search response: %v", err)
		}
		if len(result.Memories) != expected {
			t.Fatalf("%s: expected %d memories, got %d", conversationID, expected, len(result.Memories))
		}
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
