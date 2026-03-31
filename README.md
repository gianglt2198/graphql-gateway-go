# 🚀 GraphQL Federation in Go

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=for-the-badge&logo=go)
![GraphQL](https://img.shields.io/badge/GraphQL-Federation_v2-E10098?style=for-the-badge&logo=graphql)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker)

**A production-ready GraphQL Federation architecture built in Go** — featuring a unified supergraph gateway that composes multiple subgraph services through schema introspection, event-driven subscriptions, and a full observability stack.

</div>

---

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Features](#features)
- [Technology Stack](#technology-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Services & Endpoints](#services--endpoints)
- [Event-Driven Subscriptions (EDFS)](#event-driven-subscriptions-edfs)
- [Database Migrations](#database-migrations)
- [Observability & Monitoring](#observability--monitoring)
- [Deployment](#deployment)
- [Development Workflow](#development-workflow)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

`federation-go` is a comprehensive **GraphQL Federation** mono-repo demonstrating a microservices architecture with a federated GraphQL layer. It features a central gateway that automatically composes schemas from independent subgraph services, enabling clients to query across multiple domains with a single unified API.

This project is designed as a showcase of modern Go backend engineering — combining **GraphQL Federation v2**, **event-driven architecture**, **clean dependency injection**, and **full-stack observability**.

---

## Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                         Client Application                       │
└────────────────────────────┬─────────────────────────────────────┘
                             │ GraphQL Queries / Subscriptions
                             ▼
┌──────────────────────────────────────────────────────────────────┐
│                     Federation Gateway :8080                     │
│            (Schema Composer · Query Planner · Executor)          │
└────────────┬──────────────────────────────┬──────────────────────┘
             │                              │
             ▼                              ▼
┌────────────────────┐          ┌────────────────────┐
│  Account Service   │          │  Catalog Service   │
│      :8082         │          │      :8083         │
│  ─────────────     │          │  ─────────────     │
│  Users             │          │  Products          │
│  Auth / Sessions   │          │  Categories        │
└─────────┬──────────┘          └─────────┬──────────┘
          │                               │
          └───────────────┬───────────────┘
                          ▼
          ┌───────────────────────────────┐
          │         NATS :4222            │
          │  (Event Bus · EDFS Broker)    │
          └───────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
    ┌──────────┐   ┌──────────┐   ┌──────────┐
    │PostgreSQL│   │  Redis   │   │   etcd   │
    │  :5433   │   │  :6379   │   │          │
    └──────────┘   └──────────┘   └──────────┘
```

---

## ✨ Features

### 🏗️ Federation Architecture

- **GraphQL Federation v2** — Full support for `@key`, `@external`, `@requires`, `@provides`, `@shareable`, and `@link` directives
- **Dynamic Subgraph Registry** — Runtime registration and management of subgraph services
- **Automatic Schema Composition** — Supergraph generated dynamically from subgraph introspection
- **Intelligent Query Planning** — Distributes queries to the appropriate subgraphs with dependency analysis
- **Parallel Query Execution** — Independent subgraph queries execute concurrently

### 🔐 Account Service

- **User Management** — Full CRUD with relay-style cursor pagination
- **Authentication** — Register, Login, Logout, and JWT token verification
- **Session Management** — Token-based session tracking with `last_used_at` tracking
- **Soft Delete** — Users are soft-deleted, not permanently removed
- **Schema-First** — All types defined in GraphQL SDL with `gqlgen`

### 🛍️ Catalog Service

- **Product Management** — Full CRUD with name, description, price, and stock
- **Category Management** — Hierarchical product categorization
- **Many-to-Many Relations** — Products ↔ Categories with cascade delete
- **Relay Pagination** — Cursor-based pagination with filtering and ordering for all entities
- **Cross-Service Entity** — References `UserEntity` from the Account service via federation keys

### 🔄 Event-Driven Federated Subscriptions (EDFS)

- **Declarative Events** — Events defined directly in GraphQL schema via directives (`@edfs_Subscribe`, `@edfs_Publish`, `@edfs_Request`)
- **Zero Boilerplate** — No manual event publishing code in resolvers
- **NATS Integration** — Real-time event streaming over NATS message broker
- **Dynamic Topic Filtering** — Variable substitution in topic patterns (e.g., `entities.User.{userId}`)
- **CQRS Support** — Event request pattern for async command processing
- **Federation-Ready** — Events automatically include federation keys for entity resolution

### ⚡ Performance & Reliability

- **Circuit Breakers** — Fault tolerance to prevent cascading failures across subgraphs
- **Connection Pooling** — Optimized HTTP client pool for inter-service communication
- **DataLoader Pattern** — Batching entity resolution requests to eliminate N+1 queries
- **Redis Caching** — Distributed caching with configurable TTL and invalidation
- **Query Plan Caching** — Cached execution plans for repeated query patterns
- **Graceful Shutdown** — Proper lifecycle management via Uber FX hooks

### 🗄️ Database & Migrations

- **Ent ORM** — Type-safe, code-generated database access layer
- **Atlas CLI Migrations** — Versioned SQL migration files with rollback support
- **Auto-generated Repositories** — Repository pattern generated from Ent schema templates
- **Prefixed NanoIDs** — Human-readable, collision-resistant IDs (e.g., `user_abc123`, `prod_xyz789`)
- **Author Mixin** — Automatic `created_by` / `updated_by` tracking on all entities

### 🔍 Observability Stack

- **Structured Logging** — Uber Zap with log forwarding via NATS → Vector → Loki
- **Distributed Tracing** — OpenTelemetry → OTLP Collector → Grafana Tempo
- **Metrics** — Prometheus scraping with Redis cache hit/miss counters, request latency
- **Grafana Dashboards** — Unified monitoring UI for logs, traces, and metrics
- **Health Endpoints** — `/health`, `/health/live`, `/health/ready` on every service

### 🧰 Developer Experience

- **Hot Reload** — `air` live-reloading for all services during development
- **GraphQL Playground** — Embedded IDE at `/playground` for every service
- **Code Generation** — Single `make generate` regenerates GraphQL resolvers and Ent models
- **Monorepo Makefile** — Top-level `Makefile` orchestrates all services

---

## 🛠️ Technology Stack

| Category                 | Technology                                                                         |
| ------------------------ | ---------------------------------------------------------------------------------- |
| **Language**             | Go 1.24+                                                                           |
| **GraphQL**              | [gqlgen](https://gqlgen.com/) — schema-first code generation                       |
| **Federation Engine**    | [wundergraph/graphql-go-tools v2](https://github.com/wundergraph/graphql-go-tools) |
| **ORM**                  | [Ent](https://entgo.io/) — entity framework with code generation                   |
| **Dependency Injection** | [Uber FX](https://github.com/uber-go/fx)                                           |
| **Messaging**            | [NATS](https://nats.io/) — cloud-native pub/sub                                    |
| **Service Discovery**    | [etcd](https://etcd.io/) — distributed key-value store                             |
| **Caching**              | [Redis](https://redis.io/)                                                         |
| **Database**             | [PostgreSQL](https://www.postgresql.org/)                                          |
| **Migrations**           | [Atlas CLI](https://atlasgo.io/)                                                   |
| **Logging**              | [Uber Zap](https://github.com/uber-go/zap) + Loki + Vector                         |
| **Tracing**              | OpenTelemetry + Grafana Tempo                                                      |
| **Metrics**              | Prometheus + Grafana                                                               |
| **HTTP Framework**       | [Fiber](https://gofiber.io/)                                                       |
| **Containerization**     | Docker + Docker Compose                                                            |
| **Orchestration**        | Kubernetes                                                                         |
| **Live Reload**          | [Air](https://github.com/air-verse/air)                                            |

---

## 📁 Project Structure

```
federation-go/
├── package/                        # Shared internal library (monorepo package)
│   ├── common/                     # Shared constants and enums (request ID, trace ID, etc.)
│   ├── config/                     # Configuration structs (DB, NATS, Redis, JWT, etcd...)
│   ├── helpers/                    # JWT & AES encryption helpers
│   ├── infras/
│   │   ├── cache/redis/            # Redis client with OpenTelemetry metrics
│   │   ├── monitoring/
│   │   │   ├── logging/            # Zap logger with NATS & FX adapters
│   │   │   └── tracing/            # OpenTelemetry setup + Fiber/NATS middleware
│   │   ├── pubsub/nats/            # NATS client, message factory, middleware chain
│   │   └── serdes/                 # MessagePack & Gzip serializers
│   ├── modules/
│   │   ├── db/                     # Ent extensions: PNNID mixin, Author mixin, Soft Delete, Repository template
│   │   ├── graphql/                # DataLoader, EDFS schema definitions
│   │   ├── queue/                  # Asynq-based background job queue
│   │   ├── saga/                   # Saga pattern workflow engine
│   │   ├── scheduler/              # Cron-style task scheduler
│   │   └── services/
│   │       ├── graphql/
│   │       │   ├── federation/v1/  # Federation v1 manager & schema registry
│   │       │   ├── federation/v2/  # Federation v2 executor, WebSocket handler, registry
│   │       │   └── server/         # Standard GraphQL server (Fiber handler)
│   │       └── http/               # HTTP server with NATS transport
│   ├── platform/                   # Application bootstrap (Uber FX app factory)
│   └── utils/                      # NanoID, context helpers, struct conversion, async WaitGroup
│
├── services/
│   ├── gateway/                    # 🚪 Federation Router  [:8080]
│   │   ├── cmd/app/                # Entry point
│   │   ├── config/                 # Gateway-specific config loader
│   │   ├── graphql/schemas/        # Gateway-level GraphQL schemas (EDFS directives, UserEntity)
│   │   └── internal/app/           # Federation manager initialization
│   │
│   ├── account/                    # 👤 Account Subgraph  [:8082]
│   │   ├── cmd/                    # app, ent codegen, migrate entry points
│   │   ├── ent/schema/             # User & Session Ent schemas
│   │   ├── graphql/
│   │   │   ├── resolvers/          # auth, user, ent resolvers
│   │   │   └── schema/             # GraphQL SDL: auth/, user/ (query, mutation, type, filter...)
│   │   ├── internal/
│   │   │   ├── repos/              # User & Session repository implementations
│   │   │   └── services/           # UserService & AuthService business logic
│   │   └── migrations/             # Atlas SQL migration files
│   │
│   └── catalog/                    # 📦 Catalog Subgraph  [:8083]
│       ├── cmd/                    # app, ent codegen, migrate entry points
│       ├── ent/schema/             # Product & Category Ent schemas
│       ├── graphql/
│       │   ├── resolvers/          # product, category, definition resolvers
│       │   └── schema/             # GraphQL SDL: product/, category/ (query, mutation, type, filter...)
│       ├── internal/
│       │   ├── repos/              # Product & Category repository implementations
│       │   └── services/           # ProductService & CategoryService business logic
│       └── migrations/             # Atlas SQL migration files
│
├── deployments/
│   ├── configs/                    # Loki, Tempo, Prometheus, OTLP Collector, Vector configs
│   ├── docker/                     # Docker-related assets
│   └── kubernetes/                 # Kubernetes manifests
│
├── docs/
│   ├── FEDERATION.md               # Federation gateway setup guide
│   └── EVENT_BUS_GUIDE.md          # EDFS usage guide
│
├── scripts/
│   └── prd.txt                     # Product Requirements Document
│
├── docker-compose.yml              # Full local stack (DB, NATS, Redis, monitoring)
├── Dockerfile                      # Multi-stage build (supports SERVICE_NAME ARG)
├── Makefile                        # Top-level orchestration
└── package.json                    # concurrently + meta for multi-service dev
```

---

## 🚀 Getting Started

### Prerequisites

| Requirement             | Version |
| ----------------------- | ------- |
| Go                      | 1.24+   |
| Docker & Docker Compose | Latest  |
| Atlas CLI               | Latest  |
| Air (live-reload)       | Latest  |

### 1. Install Tools

```bash
# Install Go tools
go install github.com/air-verse/air@latest
go install github.com/99designs/gqlgen@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install Atlas CLI
curl -sSf https://atlasgo.sh | sh

# Install Node tools (for concurrent dev runner)
npm install
```

Or run the one-shot setup:

```bash
make setup
```

### 2. Start Infrastructure

```bash
# Starts PostgreSQL, NATS, Redis, Prometheus, Grafana, Loki, Tempo, Vector
make infra-up
```

### 3. Run Database Migrations

```bash
# From each service directory
cd services/account && make migrate-up
cd services/catalog && make migrate-up
```

### 4. Start Services

```bash
# Start all subgraph services with live-reload (account + catalog)
make run-service

# In a separate terminal, start the federation gateway
make run-gateway
```

### 5. Verify Playground SRI hash for the embedded Apollo Sandbox playground has not changed:

```
CDN_FILE=https://embeddable-sandbox.cdn.apollographql.com/06401c5415c2df085278716decfe48b0f1ba5b7c/embeddable-sandbox.umd.production.min.js
curl -s $CDN_FILE | openssl dgst -sha256 -binary | openssl base64 -A; echo
```

### 5. Access Playgrounds

| Service     | GraphQL Endpoint                | Playground                         |
| ----------- | ------------------------------- | ---------------------------------- |
| **Gateway** | `http://localhost:8080/graphql` | `http://localhost:8080/playground` |
| **Account** | `http://localhost:8082/graphql` | `http://localhost:8082/playground` |
| **Catalog** | `http://localhost:8083/graphql` | `http://localhost:8083/playground` |

---

## 🌐 Services & Endpoints

### Gateway Service `:8080`

The federation gateway composes the supergraph from all registered subgraphs. Send all client queries here.

| Endpoint            | Description                        |
| ------------------- | ---------------------------------- |
| `POST /graphql`     | Unified federated GraphQL endpoint |
| `GET /playground`   | GraphQL IDE                        |
| `GET /health`       | Aggregated health status           |
| `GET /health/live`  | Liveness probe                     |
| `GET /health/ready` | Readiness probe                    |

### Account Service `:8082`

Manages users and authentication.

**GraphQL Operations:**

```graphql
# Queries
users(after, first, before, last, orderBy, where: UserFilter): UserPaginatedConnection!
node(id: ID!): Node

# Mutations
accountAuthRegister(input: RegisterInput!): Boolean!
accountAuthLogin(input: LoginInput!): LoginEntity!
accountAuthLogout: Boolean!
accountAuthVerify(input: AuthVerifyInput!): AuthVerifyEntity!
accountCreateUser(input: CreateUserInput!): Boolean!
accountUpdateUser(id: String!, input: UpdateUserInput!): Boolean!
accountDeleteUser(id: String!): Boolean!

# Subscriptions (EDFS)
userUpdated(userID: ID!): UserEntity!
```

### Catalog Service `:8083`

Manages products and categories.

**GraphQL Operations:**

```graphql
# Queries
products(after, first, before, last, orderBy, where: ProductFilter): ProductConnection!
categories(after, first, before, last, orderBy, where: CategoryFilter): CategoryConnection!
category(id: ID!): Category!

# Mutations
catalogCreateProduct(input: CreateProductInput!): Boolean!
catalogUpdateProduct(id: String!, input: UpdateProductInput!): Boolean!
catalogDeleteProduct(id: String!): Boolean!
catalogCreateCategory(input: CreateCategoryInput!): Boolean!
catalogUpdateCategory(id: String!, input: UpdateCategoryInput!): Boolean!
```

---

## 📡 Event-Driven Federated Subscriptions (EDFS)

EDFS provides a **declarative, schema-first** approach to real-time events via GraphQL directives. No manual event publishing code is needed.

```
GraphQL Client ↔ Gateway ↔ NATS Message Broker ↔ Account / Catalog Services
                    ↓
         EDFS Directives handle all routing automatically
```

### Available Directives

| Directive         | Usage            | Description                                    |
| ----------------- | ---------------- | ---------------------------------------------- |
| `@edfs_Publish`   | On mutations     | Auto-publishes event after successful mutation |
| `@edfs_Subscribe` | On subscriptions | Subscribes to a NATS topic                     |
| `@edfs_Request`   | On mutations     | Fires async CQRS command                       |

### Example Usage

```graphql
# Schema definition — no resolver code needed
extend type Mutation {
  createUser(input: CreateUserInput!): User
    @edfs_Publish(subject: "entities.User.created")
}

extend type Subscription {
  userUpdated(userID: ID!): UserEntity!
    @edfs__natsSubscribe(subjects: ["userUpdated.{{ args.userID }}"])
}
```

For a complete guide, see [`docs/EVENT_BUS_GUIDE.md`](docs/EVENT_BUS_GUIDE.md).

---

## 🗃️ Database Migrations

Each service manages its own database schema using **Ent** + **Atlas CLI**.

### Common Commands

```bash
# Generate a new migration from Ent schema changes
make migrate-gen name=add_user_avatar

# Apply all pending migrations
make migrate-up

# Roll back to a specific version
make migrate-down version=20250628060846

# Create an empty migration for custom SQL
make migrate-new name=add_custom_indexes

# Check migration status
make migrate-status

# Regenerate migration checksums
make migrate-hash
```

For the full migration guide, see [`services/account/MIGRATION.md`](services/account/MIGRATION.md).

---

## 📊 Observability & Monitoring

The full observability stack is included in `docker-compose.yml`:

| Tool               | Port          | Purpose                                      |
| ------------------ | ------------- | -------------------------------------------- |
| **Grafana**        | `3000`        | Unified dashboards (logs + traces + metrics) |
| **Prometheus**     | `9090`        | Metrics scraping & storage                   |
| **Grafana Tempo**  | `3200`        | Distributed tracing backend                  |
| **Grafana Loki**   | `3100`        | Log aggregation                              |
| **Vector**         | —             | Log pipeline: NATS → Loki                    |
| **OTLP Collector** | `4317 / 4318` | OpenTelemetry traces & metrics ingestion     |

### Log Pipeline

```
Services → NATS (subject: logging) → Vector → Loki → Grafana
```

### Trace Pipeline

```
Services → OTLP Collector (gRPC :4317) → Tempo → Grafana
```

### Metrics Pipeline

```
Services → Prometheus (pull :9090) → Grafana
```

---

## 🐳 Deployment

### Docker

Build and run all services using Docker Compose:

```bash
# Build all service images
docker-compose build

# Start the full stack
docker-compose up -d
```

The multi-stage `Dockerfile` supports a `SERVICE_NAME` build argument:

```bash
docker build --build-arg SERVICE_NAME=account -t federation-go/account .
docker build --build-arg SERVICE_NAME=catalog -t federation-go/catalog .
docker build --build-arg SERVICE_NAME=gateway -t federation-go/gateway .
```

### Kubernetes

```bash
kubectl apply -f deployments/kubernetes/
```

---

## 🔧 Development Workflow

### Code Generation

After modifying a GraphQL schema or Ent schema:

```bash
# Inside any service directory
go generate ./...

# Or via top-level Makefile
make generate
```

### Code Quality

```bash
make lint      # Run golangci-lint on all services
make fmt       # Format Go code (goimports-reviser)
make vet       # Run go vet on all services
make tidy      # Tidy go.mod for all services
```

### Testing

```bash
go test ./...

# Or via Makefile
make test
```

### Makefile Reference

| Command            | Description                              |
| ------------------ | ---------------------------------------- |
| `make setup`       | Install all dev tools                    |
| `make infra-up`    | Start Docker infrastructure              |
| `make infra-down`  | Stop Docker infrastructure               |
| `make run-service` | Start subgraph services with hot-reload  |
| `make run-gateway` | Start federation gateway with hot-reload |
| `make build`       | Build all service binaries               |
| `make generate`    | Regenerate GraphQL & Ent code            |
| `make test`        | Run all tests                            |
| `make lint`        | Lint all services                        |
| `make clean`       | Remove generated artifacts               |

---

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'feat: add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

Please follow the existing code style and ensure all services pass `make lint` before submitting.

---

## 📄 License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

This project stands on the shoulders of outstanding open-source work. Sincere thanks to every author, maintainer, and contributor behind the following projects and ideas.

- [gqlgen](https://gqlgen.com/) — GraphQL implementation for Go
- [wundergraph/graphql-go-tools](https://github.com/wundergraph/graphql-go-tools) — High-performance GraphQL federation engine
- [Ent](https://entgo.io/) — Entity framework for Go
- [Uber FX](https://github.com/uber-go/fx) — Dependency injection framework
- [NATS](https://nats.io/) — Cloud-native messaging
- [Atlas](https://atlasgo.io/) — Database migration tool
- [Grafana OSS Stack](https://grafana.com/) — Loki, Tempo, Grafana

---

<div align="center">

⭐ If this project helped you, please consider giving it a star!

**Built with ❤️ by Liam Le — Software Engineer**

</div>
