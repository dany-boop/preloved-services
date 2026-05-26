# AntiGravity Backend — Architecture & Microservice Decisions

## 🏗️ Project Structure Overview

```
preloved-backend/
├── docker-compose.yml        # All infrastructure
├── .env.example              # Environment variables template
├── go.work                   # Go workspace (links all services)
├── Makefile                  # Dev shortcuts
├── services/                 # Each microservice (independent)
│   ├── auth-service/         # Port 8001
│   ├── user-service/         # Port 8002
│   ├── chat-service/         # Port 8003
│   ├── notification-service/ # Port 8004
│   ├── search-service/       # Port 8005
│   ├── ai-service/           # Port 8006
│   └── media-service/        # Port 8007
├── shared/                   # Code shared between services
│   ├── types/                # Common structs, response envelopes
│   ├── jwt/                  # Token generation & validation
│   ├── database/             # DB connection helpers
│   ├── rabbitmq/             # Queue publish/consume helpers
│   ├── logger/               # Zerolog setup + correlation IDs
│   └── middleware/           # Auth, CORS, correlation ID
└── infra/
    ├── krakend/              # API Gateway config
    ├── migrations/           # PostgreSQL SQL files
    └── monitoring/           # Prometheus, Grafana configs
```

---

## 🎯 Microservice Decisions

### Why each service exists as its own module

| Service | Reason for Separation | DB Used | Scales When |
|---|---|---|---|
| **auth-service** | Security isolation — auth bugs shouldn't crash other services | PostgreSQL + Redis | High login volume |
| **user-service** | Profile data changes independently from auth logic | PostgreSQL | Many profile reads |
| **chat-service** | Needs WebSocket connections, very different traffic pattern | MongoDB + Redis | Many concurrent connections |
| **notification-service** | Pure consumer — listens to RabbitMQ, no HTTP traffic | RabbitMQ consumer | High notification volume |
| **search-service** | Elasticsearch is unique infrastructure, isolated well | Elasticsearch | Heavy search load |
| **ai-service** | Expensive, slow calls (LLM APIs) — separate rate limits | PostgreSQL (pgvector) | AI usage spikes |
| **media-service** | File uploads need different timeout settings (60s vs 10s) | MinIO | Upload traffic |

---

## 🗄️ Database Decisions

### PostgreSQL — Core Business Data
**Used by:** auth-service, user-service, ai-service

**Why PostgreSQL:**
- ACID transactions — critical for user accounts, money-related operations
- Foreign keys — data integrity between users, profiles, tokens
- pgvector extension — store AI embeddings alongside your data
- Excellent Go support via `pgx`

**Tables:** users, user_profiles, refresh_tokens, email_verifications, password_resets

---

### MongoDB — Chat & Logs
**Used by:** chat-service

**Why MongoDB:**
- Chat messages = append-only, high volume, flexible format
- No joins needed — a message document contains everything
- TTL indexes — auto-delete old messages after N days
- Horizontal sharding built-in for massive scale

**Collections:** messages, rooms, room_members

---

### Redis — Cache + Real-time
**Used by:** auth-service (sessions), chat-service (pub/sub), all services (cache)

**Why Redis:**
- Fastest DB in the stack — sub-millisecond reads
- Pub/Sub for WebSocket cross-instance broadcasting
- Session storage — O(1) lookup by token
- Rate limiting counters
- Leaderboards with sorted sets (future feature)

**Keys pattern:**
```
session:{user_id}        → user session data
cache:user:{id}          → cached user profile (5 min TTL)
presence:{user_id}       → online/offline status
ratelimit:{ip}:{minute}  → API rate limit counter
```

---

### Elasticsearch — Search
**Used by:** search-service

**Why Elasticsearch:**
- Full-text search with ranking/scoring
- Fuzzy matching ("did you mean...?")
- Aggregations for faceted search filters
- Much faster than PostgreSQL `LIKE '%query%'`

**Indices:** users, content, products (future)

---

### MinIO — File Storage
**Used by:** media-service

**Why MinIO:**
- S3-compatible API — easy migration to AWS S3 in production
- Self-hosted for development
- Stores profile photos, uploaded files, documents

---

## 🔄 Service Communication Patterns

### Synchronous (HTTP via KrakenD)
```
Client → KrakenD → Service → Response
```
Use for: queries, CRUD operations, anything needing immediate response

### Asynchronous (RabbitMQ queues)
```
Service A → RabbitMQ → Service B (processes later)
```
Use for: emails, notifications, analytics, AI jobs

**Queue naming convention:**
- `email.queue` — welcome, verification, password reset
- `notification.queue` — push notifications, in-app alerts
- `analytics.queue` — user events, tracking
- `ai.tasks.queue` — async AI inference jobs

---

## 📦 Each Service Internal Structure

```
services/auth-service/
├── cmd/
│   └── main.go          ← Entry point: loads config, starts server
├── config/
│   └── config.go        ← Reads .env, type-safe config struct
├── internal/            ← All internal code (can't be imported externally)
│   ├── handler/         ← HTTP handlers (thin: parse request, call service)
│   ├── service/         ← Business logic (all the real work)
│   ├── repository/      ← Database queries (SQL/Mongo calls only here)
│   └── model/           ← Database model structs
└── go.mod               ← Own dependencies
```

**The rule: handler calls service, service calls repository. Never skip layers.**

---

## 🚀 Getting Started

```bash
# 1. Clone and setup
cp .env.example .env
# Edit .env with your values

# 2. Start all infrastructure
make dev-infra

# 3. Run migrations
make migrate

# 4. Start a service
make run-auth

# 5. Test the API
curl http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","username":"testuser"}'
```

---

## 🔑 Key Tools & Why

| Tool | Purpose | Learn It |
|---|---|---|
| **Gin** | HTTP framework — fast, simple, great for beginners | [gin-gonic.com](https://gin-gonic.com) |
| **pgx** | PostgreSQL driver — better than database/sql | [jackc/pgx](https://github.com/jackc/pgx) |
| **zerolog** | Structured JSON logging | [rs/zerolog](https://github.com/rs/zerolog) |
| **gorilla/websocket** | WebSocket in Go | [gorilla/websocket](https://github.com/gorilla/websocket) |
| **golang-jwt** | JWT creation/parsing | [golang-jwt](https://github.com/golang-jwt/jwt) |
| **amqp091-go** | RabbitMQ official Go client | [rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go) |
| **KrakenD** | API Gateway — handles auth, rate limiting, routing | [krakend.io](https://www.krakend.io) |
