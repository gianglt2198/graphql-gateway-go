# Event Bus Guide - EDFS (Event Driven Federated Subscriptions)

## Overview

This guide demonstrates how to implement **Event Driven Federated Subscriptions (EDFS)** using GraphQL directives and NATS as the message transport. EDFS provides a declarative, schema-first approach to real-time events in GraphQL federation architecture.

### Key Benefits of EDFS

✅ **Declarative** - Events defined in GraphQL schema using directives  
✅ **Automatic** - No manual event handling code required  
✅ **Type-safe** - Events tied to GraphQL operations and types  
✅ **Federation-ready** - Automatic federation key handling  
✅ **Clean** - No centralized event definitions needed  
✅ **Dynamic** - Topic filtering and variable substitution  

## EDFS Architecture

```
GraphQL Client ↔ Gateway (Router) ↔ NATS Message Broker ↔ Services (Account, Catalog)
                    ↓
              EDFS Directives handle all event routing automatically
```

Instead of manually writing event publishers and subscribers, you simply add directives to your GraphQL schema:

- `@eventPublish` - Automatically publishes events after successful mutations
- `@eventSubscribe` - Automatically subscribes to NATS topics for subscriptions  
- `@eventRequest` - Triggers event requests for command patterns
- `@openfed_subscriptionFilter` - Enables dynamic topic filtering

## Configuration

### 1. NATS Server Setup

```yaml
# docker-compose.yml
services:
  nats:
    image: nats:latest
    ports:
      - "4223:4222"
    command: ["-js", "-m", "8222"]
```

### 2. Service Configuration

```yaml
# services/account/config.yml
nats:
  endpoints:
    - "nats://localhost:4223"
  
# services/gateway/config.yml  
nats:
  endpoints:
    - "nats://localhost:4223"
```

## EDFS Directives Usage

### Basic Event Publishing

```graphql
extend type Mutation {
  # Automatically publishes to "entities.User.created" after successful mutation
  createUser(input: CreateUserInput!): User 
    @eventPublish(topic: "entities.User.created", entityKey: "id")
    
  # Automatically publishes to "entities.User.updated" 
  updateUser(id: ID!, input: UpdateUserInput!): User 
    @eventPublish(topic: "entities.User.updated", entityKey: "id")
}
```

### Event Subscriptions

```graphql
extend type Subscription {
  # Subscribes to all User entity events
  userEvents: User 
    @eventSubscribe(topic: "entities.User.*", sourceName: "account")
    
  # Dynamic subscription with filtering
  userEventsForId(userId: ID!): User
    @eventSubscribe(topic: "entities.User.{userId}", sourceName: "account")
    @openfed_subscriptionFilter(condition: "{userId}")
}
```

### Event Requests (CQRS Pattern)

```graphql
extend type Mutation {
  # Triggers event request for async processing
  deleteUser(id: ID!): Boolean
    @eventRequest(topic: "services.account.requests", sourceName: "account")
}
```

## Complete Schema Example

### Account Service (`services/account/graphql/schema/user/mutation.gql`)

```graphql
# EDFS directives are automatically included by federation manager

extend type Mutation {
  createUser(input: CreateUserInput!): User 
    @eventPublish(topic: "entities.User.created", entityKey: "id")
    
  updateUser(id: ID!, input: UpdateUserInput!): User 
    @eventPublish(topic: "entities.User.updated", entityKey: "id")
    
  authenticateUser(email: String!, password: String!): AuthPayload
    @eventPublish(topic: "entities.User.authenticated", entityKey: "user.id")
}

extend type Subscription {
  userEvents: User 
    @eventSubscribe(topic: "entities.User.*", sourceName: "account")
}
```

### Gateway Service (`services/gateway/graphql/schema/federation.gql`)

```graphql
extend type Subscription {
  # Gateway subscribes to all entity events for real-time federation
  allEntityEvents: EntityEvent
    @eventSubscribe(topic: "entities.*", sourceName: "gateway")
    
  # Service health monitoring
  serviceHealth: ServiceStatus
    @eventSubscribe(topic: "services.*.health", sourceName: "monitoring")
}

type EntityEvent {
  id: ID!
  __typename: String!
  eventType: String!
  timestamp: String!
  data: String!
}
```

## Event Topic Patterns

EDFS uses consistent topic naming patterns:

### Entity Events
```
entities.User.created      # User creation events
entities.User.updated      # User update events  
entities.User.{userId}     # User-specific events (dynamic filtering)
entities.Product.*         # All Product events
entities.*                 # All entity events
```

### Service Events  
```
services.account.health    # Account service health
services.catalog.requests  # Catalog service requests
services.*.health          # All service health events
```

### Gateway Events
```
gateway.requests           # Gateway request events
gateway.responses          # Gateway response events
gateway.errors             # Gateway error events
```

## Client Usage

### Subscription Example

```javascript
// GraphQL subscription automatically handled by EDFS
const USER_EVENTS_SUBSCRIPTION = gql`
  subscription UserEvents {
    userEvents {
      id
      email
      name
      __typename
    }
  }
`;

// Use with any GraphQL client (Apollo, Relay, etc.)
const { data } = useSubscription(USER_EVENTS_SUBSCRIPTION);
```

### Mutation with Automatic Event Publishing

```javascript
const CREATE_USER_MUTATION = gql`
  mutation CreateUser($input: CreateUserInput!) {
    createUser(input: $input) {
      id
      email
      name
    }
  }
`;

// Event is automatically published to "entities.User.created" via @eventPublish directive
const [createUser] = useMutation(CREATE_USER_MUTATION);
```

## Advanced Features

### Dynamic Topic Filtering

```graphql
extend type Subscription {
  # Only receives events for specific user ID
  userEventsForId(userId: ID!): User
    @eventSubscribe(topic: "entities.User.{userId}", sourceName: "account")
    @openfed_subscriptionFilter(condition: "{userId}")
}
```

### Federation Compatibility

Events automatically include federation keys:

```json
{
  "id": "user_123",
  "__typename": "User", 
  "eventType": "created",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "id": "user_123",
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

### CQRS Pattern Implementation

```graphql
# Command (triggers async processing)
extend type Mutation {
  processLargeDataset(input: DatasetInput!): JobID
    @eventRequest(topic: "services.processing.jobs", sourceName: "api")
}

# Query (read results)
extend type Subscription {
  jobProgress(jobId: ID!): JobProgress
    @eventSubscribe(topic: "jobs.{jobId}.progress", sourceName: "processing")
    @openfed_subscriptionFilter(condition: "{jobId}")
}
```

## Implementation Benefits

### Compared to Manual Event Handling

**EDFS Approach:**
```graphql
# Simply add directive to schema
createUser(input: CreateUserInput!): User 
  @eventPublish(topic: "entities.User.created", entityKey: "id")
```

**Manual Approach:**
```go
// Requires custom event handling code in resolver
func (r *mutationResolver) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
    user, err := r.service.CreateUser(ctx, input)
    if err != nil {
        return nil, err
    }
    
    // Manual event publishing
    r.eventBus.Publish(ctx, "entities.User.created", map[string]interface{}{
        "id": user.ID,
        "__typename": "User",
        "data": user,
    })
    
    return user, nil
}
```

### Schema-First Development

1. **Define events in GraphQL schema** using directives
2. **Generate resolvers** with gqlgen (directives are automatically handled)
3. **No event handling code** needed in business logic
4. **Federation-ready** events with automatic key extraction

## Migration from Manual Events

If you have existing manual event handling code, you can migrate to EDFS:

1. **Add directives** to your GraphQL schema
2. **Remove manual event publishing** from resolvers  
3. **Update subscriptions** to use `@eventSubscribe`
4. **Test** that events still flow correctly

The EDFS system is backward compatible and can coexist with manual event handling during migration.

## Troubleshooting

### Directive Not Working

1. Ensure NATS is running and accessible
2. Check that EventBus is provided to FederationManager
3. Verify directive syntax in GraphQL schema
4. Check logs for EDFS initialization messages

### Events Not Received

1. Verify topic patterns match between publish and subscribe
2. Check NATS connection in service logs
3. Ensure subscription is active in GraphQL client
4. Test with simple topic pattern first

### Performance Considerations

1. Use specific topic patterns to reduce unnecessary events
2. Implement subscription filtering for large event volumes
3. Consider event batching for high-throughput scenarios
4. Monitor NATS metrics and connection pools

## Best Practices

1. **Use semantic topic patterns** - `entities.Type.action` format
2. **Include federation keys** - Always specify `entityKey` for entity events
3. **Filter subscriptions** - Use specific topics and filtering to reduce noise
4. **Handle errors gracefully** - Events are fire-and-forget by design
5. **Monitor performance** - Track event throughput and subscription counts
6. **Version your events** - Plan for schema evolution in event payloads

---

EDFS provides a powerful, declarative approach to real-time GraphQL subscriptions that eliminates the need for manual event handling while maintaining full federation compatibility and type safety. 