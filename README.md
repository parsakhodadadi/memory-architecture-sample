# MemoryArchitectureSample

A concise Go demonstration of chatbot memory architecture:

- Redis for short-term conversation memory
- PostgreSQL with pgvector for long-term semantic memory
- Separate memory modules behind small interfaces
- Time-based retention for both memory layers
- Swagger UI for testing the HTTP API

The chatbot uses a deterministic local response generator, so no paid LLM API
or API key is required.

Full setup and usage instructions are added with the Docker task.
