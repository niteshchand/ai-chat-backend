# AI Chat Backend

A production-grade AI chat backend built with **Golang (Gin)** featuring RAG (Retrieval Augmented Generation), streaming responses, JWT authentication, and rate limiting.

🌐 **Live Demo:** [ai-chat-frontend-iota.vercel.app](https://ai-chat-frontend-iota.vercel.app)

---

## Features

- **Streaming AI responses** — Server-Sent Events (SSE) delivering tokens in real time
- **RAG Pipeline** — Upload PDFs, generate embeddings, semantic search, document-grounded answers
- **JWT Authentication** — Secure register/login with bcrypt password hashing
- **Multi-conversation** — Multiple independent chat threads per user
- **Custom system prompts** — Per-conversation AI personality
- **Rate limiting** — Sliding window algorithm (20 requests/hour per user)
- **Vector embeddings** — Cosine similarity search using Google Gemini embeddings
- **PostgreSQL** — Persistent chat history, documents, and users

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Golang |
| Framework | Gin |
| Database | PostgreSQL + GORM |
| AI/LLM | Groq API (llama-3.3-70b) |
| Embeddings | Google Gemini (text-embedding-001) |
| Auth | JWT + bcrypt |
| Deployment | Render |

---

## Architecture

```
┌─────────────────────────────────────────┐
│              Next.js Frontend           │
└──────────────────┬──────────────────────┘
                   │ HTTP / SSE
┌──────────────────▼──────────────────────┐
│           Gin Router + Middleware        │
│         JWT Auth + Rate Limiting         │
└──┬──────────────┬────────────────┬──────┘
   │              │                │
┌──▼───┐    ┌────▼────┐    ┌──────▼─────┐
│ Auth │    │  Chat   │    │  RAG / PDF │
│Handler│   │ Handler │    │  Handler   │
└──────┘    └────┬────┘    └──────┬─────┘
                 │                │
┌────────────────▼────────────────▼──────┐
│              Services                   │
│  Groq API │ Embeddings │ Chunker        │
└────────────────────────────────────────┘
                 │
┌────────────────▼───────────────────────┐
│             PostgreSQL                  │
│  users │ conversations │ messages       │
│  documents │ document_chunks            │
└────────────────────────────────────────┘
```

---

## RAG Pipeline

```
INGESTION:
PDF Upload → Text Extraction → 500-token Chunks
→ Gemini Embeddings → Store in PostgreSQL

QUERY:
User Question → Embed Question
→ Cosine Similarity Search → Top 3 Chunks
→ Inject as Context → LLM → Grounded Answer
```

---

## Project Structure

```
ai-chat-backend/
├── main.go
├── db/
│   └── db.go              # Database connection + migrations
├── handlers/
│   ├── auth.go            # Register + Login
│   ├── chat.go            # Non-streaming chat
│   ├── stream.go          # SSE streaming + RAG injection
│   ├── conversation.go    # CRUD conversations
│   ├── history.go         # Message history
│   ├── upload.go          # PDF upload + chunking + embedding
│   └── search.go          # Vector similarity search
├── services/
│   ├── groq.go            # Groq API + streaming
│   ├── embeddings.go      # Gemini embeddings + cosine similarity
│   ├── chunker.go         # Text chunking with overlap
│   ├── retrieval.go       # RAG retrieval pipeline
│   └── jwt.go             # Token generation + verification
├── middleware/
│   ├── auth.go            # JWT middleware
│   └── ratelimit.go       # Sliding window rate limiter
└── models/
    ├── user.go
    ├── message.go
    ├── conversation.go
    └── document.go
```

---

## API Endpoints

### Auth
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/register` | Create account → returns JWT |
| POST | `/login` | Login → returns JWT |

### Conversations (Protected)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/conversations` | List all conversations |
| POST | `/conversations` | Create conversation |
| GET | `/conversations/:id` | Get conversation + messages |
| PUT | `/conversations/:id` | Update title / system prompt |
| DELETE | `/conversations/:id` | Delete conversation |

### Chat (Protected + Rate Limited)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/chat` | Standard chat response |
| POST | `/chat/stream` | SSE streaming response with RAG |

### RAG
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/upload` | Upload PDF → extract → embed |
| POST | `/search` | Semantic search over document chunks |
| GET | `/usage` | Current rate limit usage |

---

## Getting Started

### Prerequisites
- Go 1.21+
- PostgreSQL
- Groq API key → [console.groq.com](https://console.groq.com)
- Gemini API key → [aistudio.google.com](https://aistudio.google.com)

### Setup

```bash
# Clone
git clone https://github.com/niteshchand/ai-chat-backend.git
cd ai-chat-backend

# Install dependencies
go mod tidy

# Create .env
cp .env.example .env
# Fill in your keys

# Run
go run main.go
```

### Environment Variables

```env
GROQ_API_KEY=your_groq_api_key
GEMINI_API_KEY=your_gemini_api_key
JWT_SECRET=your_super_secret_key
DATABASE_URL=postgres://user:password@localhost:5432/ai_chat?sslmode=disable
```

---

## Key Technical Decisions

**Why Go for AI backend?**
Go's concurrency model handles SSE streaming efficiently. Goroutines make it easy to manage long-lived connections without blocking. Significantly faster than Node.js for CPU-bound tasks like cosine similarity calculations.

**Why cosine similarity in Go instead of pgvector?**
Avoids external vector DB dependency. Cosine similarity over 8-100 chunks is fast enough in pure Go. Simpler infrastructure — one PostgreSQL instance handles everything.

**Why sliding window rate limiting?**
More accurate than fixed window. Prevents edge-case bursts at window boundaries. Easy to explain in production incidents.

---

## License

MIT
