package httpapi

var swaggerHTML = []byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Memory Architecture API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({url: "/openapi.json", dom_id: "#swagger-ui"});
  </script>
</body>
</html>`)

var openAPISpec = []byte(`{
  "openapi": "3.0.3",
  "info": {
    "title": "Memory Architecture Sample API",
    "version": "1.0.0",
    "description": "A testable chatbot memory API built with clean architecture."
  },
  "servers": [{"url": "/"}],
  "paths": {
    "/health": {
      "get": {
        "summary": "Check API health",
        "responses": {
          "200": {
            "description": "API is healthy",
            "content": {"application/json": {"schema": {"$ref": "#/components/schemas/Health"}}}
          }
        }
      }
    },
    "/api/v1/chat": {
      "post": {
        "summary": "Send a message",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {"$ref": "#/components/schemas/ChatRequest"},
              "example": {"conversationId": "demo-1", "message": "My favorite language is Go."}
            }
          }
        },
        "responses": {
          "200": {
            "description": "Chat reply",
            "content": {"application/json": {"schema": {"$ref": "#/components/schemas/ChatResponse"}}}
          },
          "400": {"description": "Invalid request"}
        }
      }
    },
    "/api/v1/memories/search": {
      "post": {
        "summary": "Search long-term semantic memory",
        "description": "Creates a local embedding for the query and uses pgvector cosine similarity.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {"$ref": "#/components/schemas/MemorySearchRequest"},
              "example": {"conversationId": "demo-1", "query": "Which language do I like?", "limit": 5}
            }
          }
        },
        "responses": {
          "200": {"description": "Related semantic memories"},
          "400": {"description": "Invalid request"}
        }
      }
    },
    "/api/v1/memories/{memoryId}": {
      "delete": {
        "summary": "Delete one long-term memory",
        "parameters": [
          {
            "name": "memoryId",
            "in": "path",
            "required": true,
            "description": "The id returned by semantic memory search.",
            "schema": {"type": "string"}
          }
        ],
        "responses": {
          "204": {"description": "Memory deleted"},
          "404": {"description": "Memory not found"}
        }
      }
    },
    "/api/v1/conversations/{conversationId}/messages": {
      "get": {
        "summary": "Get recent conversation messages",
        "parameters": [
          {"name": "conversationId", "in": "path", "required": true, "schema": {"type": "string"}},
          {"name": "limit", "in": "query", "schema": {"type": "integer", "default": 20, "minimum": 1, "maximum": 100}}
        ],
        "responses": {"200": {"description": "Conversation history"}}
      },
      "delete": {
        "summary": "Clear conversation messages",
        "parameters": [
          {"name": "conversationId", "in": "path", "required": true, "schema": {"type": "string"}}
        ],
        "responses": {"204": {"description": "Conversation cleared"}}
      }
    },
    "/api/v1/conversations/{conversationId}/memories": {
      "delete": {
        "summary": "Delete all long-term memories for a conversation",
        "parameters": [
          {"name": "conversationId", "in": "path", "required": true, "schema": {"type": "string"}}
        ],
        "responses": {"204": {"description": "Long-term memories deleted"}}
      }
    }
  },
  "components": {
    "schemas": {
      "Health": {
        "type": "object",
        "properties": {"status": {"type": "string", "example": "ok"}}
      },
      "Message": {
        "type": "object",
        "properties": {
          "id": {"type": "string"},
          "conversationId": {"type": "string"},
          "role": {"type": "string", "enum": ["user", "assistant"]},
          "content": {"type": "string"},
          "createdAt": {"type": "string", "format": "date-time"}
        }
      },
      "ChatRequest": {
        "type": "object",
        "required": ["conversationId", "message"],
        "properties": {
          "conversationId": {"type": "string"},
          "message": {"type": "string"}
        }
      },
      "ChatResponse": {
        "type": "object",
        "properties": {
          "conversationId": {"type": "string"},
          "reply": {"type": "string"},
          "context": {"type": "array", "items": {"$ref": "#/components/schemas/Message"}},
          "recalledMemory": {"type": "array", "items": {"$ref": "#/components/schemas/SemanticMemory"}}
        }
      },
      "MemorySearchRequest": {
        "type": "object",
        "required": ["conversationId", "query"],
        "properties": {
          "conversationId": {"type": "string"},
          "query": {"type": "string"},
          "limit": {"type": "integer", "default": 5, "minimum": 1, "maximum": 20}
        }
      },
      "SemanticMemory": {
        "type": "object",
        "properties": {
          "id": {"type": "string"},
          "conversationId": {"type": "string"},
          "content": {"type": "string"},
          "similarity": {"type": "number", "format": "double"},
          "createdAt": {"type": "string", "format": "date-time"},
          "expiresAt": {"type": "string", "format": "date-time"}
        }
      }
    }
  }
}`)
