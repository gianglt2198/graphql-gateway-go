package gateway

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gianglt2198/graphql-gateway-go/gateway/internal/etcd"
	"github.com/gianglt2198/graphql-gateway-go/gateway/internal/federation"
	"github.com/gianglt2198/graphql-gateway-go/gateway/internal/handler"
	"github.com/gianglt2198/graphql-gateway-go/gateway/internal/registry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wundergraph/graphql-go-tools/execution/engine"
	"github.com/wundergraph/graphql-go-tools/execution/graphql"
)

type Server struct {
	registry   *registry.SchemaRegistry
	federation *federation.FederationManager
	router     *http.ServeMux
	handler    http.Handler
}

func NewServer(etcdClient *etcd.Client, router *http.ServeMux) (*Server, error) {

	var gqlHandlerFactory handler.HandlerFactoryFn = func(schema *graphql.Schema, engine *engine.ExecutionEngine) http.Handler {
		return handler.NewGraphqlHTTPHandler(schema, engine)
	}

	registry := registry.NewSchemaRegistry(etcdClient)
	federation := federation.NewFederationManager(registry, gqlHandlerFactory)

	server := &Server{
		registry:   registry,
		federation: federation,
		router:     router,
	}

	// Set up routes
	server.setupRoutes()

	// Create initial schema (may fail if no services are registered yet)
	if err := server.refreshSchema(); err != nil {
		log.Printf("Warning: Initial schema creation failed: %v", err)
	}

	// Set up schema refresh on changes
	go server.watchSchemaChanges()

	return server, nil
}

// setupRoutes sets up the HTTP routes
func (s *Server) setupRoutes() {
	// GraphQL endpoint
	s.router.HandleFunc("/graphql", s.graphqlHandler)

	// Health check
	s.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			fmt.Println("error written response", err)
		}
	})

	// Metrics
	s.router.Handle("/metrics", promhttp.Handler())
}

// graphqlHandler handles GraphQL requests
func (s *Server) graphqlHandler(w http.ResponseWriter, r *http.Request) {
	if s.handler == nil {
		http.Error(w, "GraphQL schema not ready", http.StatusServiceUnavailable)
		return
	}

	s.handler.ServeHTTP(w, r)
}

// refreshSchema refreshes the federated schema
func (s *Server) refreshSchema() error {
	if err := s.federation.RefreshFederation(); err != nil {
		log.Println("Schema refresh failed")
		return err
	}

	srv := s.federation.GetHandler()

	s.handler = srv

	log.Println("Schema refreshed successfully")
	return nil
}

// watchSchemaChanges watches for schema changes and refreshes the schema
func (s *Server) watchSchemaChanges() {
	// Listen for global schema changes
	changes, unregister := s.registry.RegisterSchemaListener("*")
	defer unregister()

	// Debounce schema updates to avoid rapid refreshes
	var lastUpdate time.Time
	const debounceInterval = 2 * time.Second

	for range changes {
		now := time.Now()
		if now.Sub(lastUpdate) < debounceInterval {
			// Skip if too soon after last update
			continue
		}

		lastUpdate = now
		if err := s.refreshSchema(); err != nil {
			log.Printf("Error refreshing schema: %v", err)
		}
	}
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
