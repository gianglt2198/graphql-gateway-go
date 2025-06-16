# GraphQL Federation in Go

A comprehensive GraphQL Federation architecture built in Go, featuring a super-graph that gathers all sub-graphs through introspection and creates modular services for common functionality.

## Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Gateway     │    │   Sub-graphs    │    │   Aggregator    │
│   (Federation   │◄──►│   (Services)    │◄──►│  (Federation    │
│    Router)      │    │                 │    │   Composer)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │      NATS       │
                    │   (Messaging)   │
                    └─────────────────┘
```
### Core Components
- **Gateway Service**: Federation router that directs queries to appropriate services
- **Aggregator Service**: Federation composer that combines schemas and resolves entities
- **Subgraph Services**:
  - **Account Service**: Manages user accounts and authentication
  - **Catalog Service**: Handles product catalog and inventory

## Technology Stack

- **Go 1.22+**: Modern Go features and patterns
- **GraphQL Federation**: Using gqlgen for schema-first development
- **Ent ORM**: For database access and code generation
- **Uber FX**: For dependency injection and application lifecycle
- **NATS**: For messaging and event-driven communication
- **etcd**: For service discovery and configuration
- **Redis**: For caching and distributed state
- **PostgreSQL**: For persistent storage
- **Prometheus & Zap**: For metrics and structured logging
- **Docker & Kubernetes**: For containerization and orchestration

## Project Structure

```
federation-go/
├── package/                    # Shared packages and modules
│   ├── common/                 # Common utilities and constants
│   ├── config/                 # Configuration management
│   ├── infras/                 # Infrastructure components
│   │   ├── cache/              # Redis cache implementation
│   │   ├── monitoring/         # Metrics and logging
│   │   ├── pubsub/             # NATS messaging
│   │   └── serdes/             # Serialization/deserialization
│   ├── modules/                # Shared business modules
│   │   └── db/                 # Database utilities and mixins
│   ├── platform/               # Platform components
│   └── utils/                  # Utility functions
├── services/                   # Service implementations
│   ├── gateway/                # Federation router
│   ├── aggregator/             # Federation composer
│   ├── account/                # Account service
│   └── catalog/                # Catalog service
├── deployments/                # Deployment configurations
│   ├── docker/                 # Docker-related files
│   └── kubernetes/             # Kubernetes manifests
└── docs/                       # Documentation
```

## Getting Started

### Prerequisites

- Go 1.24.1 or later
- Docker and Docker Compose
- PostgreSQL
- NATS
- etcd
- Redis

### Setup and Installation

### Quick Start

1. **Clone and setup**:
   ```bash
   git clone <repository-url>
   cd federation-go
   make setup
   ```

2. **Start infrastructure**:
   ```bash
   make infra-up
   ```

3. **Run the services**
   ```bash
   make mod-tidy
   make run-service
   ```

4. **Access the federation gateway**:
   ```
   make run-gateway
   ```

### Development Workflow

1. **Generate GraphQL code**

   After modifying GraphQL schemas, regenerate the code:

   ```bash
   cd services/<service-name>
   go generate ./...
   ```

2. **Generate Ent models**

   After modifying Ent schemas, regenerate the models:

   ```bash
   make generate
   ```

3. **Run database migrations**

   ```bash
   make migrate-up
   ```

## Health Checks

All services expose health check endpoints:

- `/health` - Comprehensive health status with detailed checks
- `/health/live` - Simple liveness probe
- `/health/ready` - Readiness probe

## API Documentation

### GraphQL Endpoints

- Gateway: `http://localhost:8080/graphql`
- Account Service: `http://localhost:8082/graphql`
- Catalog Service: `http://localhost:8083/graphql`

### GraphQL Playground

GraphQL Playground is available at:

- Gateway: `http://localhost:8080/playground`
- Account Service: `http://localhost:8082/playground`
- Catalog Service: `http://localhost:8083/playground`

## Testing

Run tests with:

```bash
go test ./...
```

Or use the Makefile:

```bash
make test
```

## Deployment

### Docker

Build and run with Docker:

```bash
docker-compose build
docker-compose up
```

### Kubernetes

Deploy to Kubernetes:

```bash
kubectl apply -f deployments/kubernetes/
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [gqlgen](https://gqlgen.com/) - GraphQL implementation for Go
- [Ent](https://entgo.io/) - Entity framework for Go
- [Uber FX](https://github.com/uber-go/fx) - Dependency injection framework
- [NATS](https://nats.io/) - Cloud native messaging system
- [etcd](https://etcd.io/) - Distributed key-value store 