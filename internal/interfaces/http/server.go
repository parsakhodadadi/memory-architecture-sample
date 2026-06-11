package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/example/memory-architecture-sample/internal/application/chat"
)

type Server struct {
	chat *chat.Service
}

func NewServer(chatService *chat.Service) *Server {
	return &Server{chat: chatService}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("POST /api/v1/chat", s.sendMessage)
	mux.HandleFunc("GET /api/v1/conversations/{conversationID}/messages", s.history)
	mux.HandleFunc("DELETE /api/v1/conversations/{conversationID}/messages", s.clear)
	mux.HandleFunc("GET /openapi.json", s.openAPI)
	mux.HandleFunc("GET /swagger", s.swagger)
	mux.HandleFunc("GET /swagger/", s.swagger)
	return recoverMiddleware(jsonContentType(mux))
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) sendMessage(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ConversationID string `json:"conversationId"`
		Message        string `json:"message"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	output, err := s.chat.Send(r.Context(), chat.SendInput{
		ConversationID: request.ConversationID,
		Message:        request.Message,
	})
	if err != nil {
		s.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, output)
}

func (s *Server) history(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	messages, err := s.chat.History(r.Context(), r.PathValue("conversationID"), limit)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"messages": messages})
}

func (s *Server) clear(w http.ResponseWriter, r *http.Request) {
	if err := s.chat.Clear(r.Context(), r.PathValue("conversationID")); err != nil {
		s.writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) openAPI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(openAPISpec)
}

func (s *Server) swagger(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(swaggerHTML)
}

func (s *Server) writeServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, chat.ErrConversationIDRequired) || errors.Is(err, chat.ErrMessageRequired) {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, "internal_error", "the request could not be completed")
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		return errors.New("Content-Type must be application/json")
	}

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errors.New("body must contain valid JSON with only supported fields")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func jsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
