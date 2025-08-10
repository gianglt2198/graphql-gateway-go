# Federation Gateway Setup

This document explains how to use the **GraphQL Federation Gateway** to orchestrate the **Account** and **Catalog** microservices.

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client App    │    │ Federation      │    │   Subgraphs     │
│                 │    │    Gateway      │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │   GraphQL   │◄├────┤►│  Composer   │◄├────┤►│   Account   │ │
│ │   Queries   │ │    │ │   Planner   │ │    │ │ :8083/graphql│ │
│ └─────────────┘ │    │ │  Executor   │ │    │ └─────────────┘ │
└─────────────────┘    │ └─────────────┘ │    │ ┌─────────────┐ │
                       │                 │    │ │   Catalog   │ │
                       │   :8080/graphql │    │ │ :8084/graphql│ │
                       └─────────────────┘    │ └─────────────┘ │
                                              └─────────────────┘
```

## 🚀 Quick Start

### Step 1: Start Subgraph Services

**Terminal 1 - Account Service:**
```bash
cd services/account
go run cmd/app/main.go
```

**Terminal 2 - Catalog Service:**
```bash
cd services/catalog  
go run cmd/app/main.go
```

### Step 2: Start Federation Gateway

**Terminal 3 - Federation Gateway:**
```bash
cd services/gateway
go run cmd/app/main.go
```

### Step 3: Test the Setup

**Option A - Use the test script:**
```bash
./test-federation.sh
```

**Option B - Manual testing:**

1. **Check individual services:**
   - Account: http://localhost:8083/playground
   - Catalog: http://localhost:8084/playground

2. **Access federation gateway:**
   - Gateway: http://localhost:8080/playground

## 🔧 Configuration

The federation gateway is configured in `services/gateway/config.yml`:

```yaml
federation:
  enabled: true
  subgraphs:
    - name: "account"
      url: "http://localhost:8083/graphql"
      timeout: 30
      retries: 3
    - name: "catalog" 
      url: "http://localhost:8084/graphql"
      timeout: 30
      retries: 3
```

## 🧪 Testing Federation

### Health Check Query
```graphql
query HealthCheck {
  __schema {
    queryType {
      name
    }
  }
}
```

### Service Discovery
```graphql
query ServiceInfo {
  _service {
    sdl
  }
}
```

## 📊 Features Implemented

### ✅ Phase 1: Foundation
- [x] **Subgraph Registry**: Dynamic registration and management
- [x] **Schema Composition**: Federated schema generation
- [x] **Query Planning**: Query decomposition and planning
- [x] **Configuration**: YAML-based setup with hot reload

### ✅ Phase 2: Query Execution  
- [x] **Data Source Manager**: HTTP client management with connection pooling
- [x] **Circuit Breakers**: Fault tolerance for subgraph failures
- [x] **Request Batching**: Optimized execution for multiple requests
- [x] **Performance Metrics**: Latency tracking and success rates
- [x] **Health Monitoring**: Automatic subgraph health checks

## 🎯 Current Capabilities

### Federation Gateway
- **Port**: 8080
- **Endpoints**: 
  - GraphQL: `/graphql`
  - Playground: `/playground`
- **Features**:
  - Subgraph registration from config
  - Health monitoring every 30 seconds
  - Circuit breaker protection
  - Request/response logging
  - Connection pooling

### Account Service Integration
- **Port**: 8083
- **Capabilities**: User management, authentication
- **Federation**: Registered as "account" subgraph

### Catalog Service Integration  
- **Port**: 8084
- **Capabilities**: Product and category management
- **Federation**: Registered as "catalog" subgraph

## 🔍 Monitoring & Debugging

### Logs
The gateway provides detailed logs for:
- Subgraph registration
- Health check results
- Query execution steps
- Circuit breaker state changes
- Performance metrics

### Health Monitoring
```bash
curl http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"query { __schema { queryType { name } } }"}'
```

### Metrics
- Request latency
- Success/failure rates  
- Circuit breaker trips
- Connection pool statistics

## 🚧 Next Steps (Phase 3+)

### Advanced Features
- [ ] **Parallel Execution**: Concurrent subgraph requests
- [ ] **Query Optimization**: Intelligent batching and caching
- [ ] **Subscription Support**: Real-time data streaming
- [ ] **Advanced Caching**: Redis-based result caching

### Production Features
- [ ] **Authentication**: JWT token forwarding
- [ ] **Rate Limiting**: Request throttling
- [ ] **Distributed Tracing**: Request correlation
- [ ] **Security**: CORS, CSP, input validation

## 🐛 Troubleshooting

### Common Issues

**Q: Gateway fails to start**
- Check that account and catalog services are running
- Verify config.yml has correct URLs and ports
- Check logs for specific error messages

**Q: Subgraph not registering**  
- Ensure subgraph GraphQL endpoint is accessible
- Check network connectivity between services
- Verify subgraph responds to introspection queries

**Q: Queries failing**
- Check subgraph health in gateway logs
- Verify query syntax is valid GraphQL
- Monitor circuit breaker state

---

**Your federation gateway is ready to orchestrate Account and Catalog services!** 🎉 