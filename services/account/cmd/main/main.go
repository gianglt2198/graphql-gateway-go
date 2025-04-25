package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/gianglt2198/graphql-gateway-go/pkg/etcd"
	"github.com/gianglt2198/graphql-gateway-go/services/account/graph/generated"
	"github.com/gianglt2198/graphql-gateway-go/services/account/graph/resolvers"
	"github.com/gianglt2198/graphql-gateway-go/services/account/internal/config"
)

func main() {

	// Init
	cfg := config.MustLoadConfig()

	// Create GraphQL server
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers.Resolver{}}))

	srv.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		graphql.GetOperationContext(ctx).DisableIntrospection = false
		return next(ctx)
	})

	// Set up HTTP handlers
	http.Handle("/", playground.Handler("GraphQL playground", "/graphql"))
	http.Handle("/graphql", srv)

	// Generate service URLs (in production this would be configurable)

	// Start HTTP server in a goroutine
	server := &http.Server{Addr: fmt.Sprintf(":%d", cfg.App.Port)}
	go func() {
		log.Printf("Account service starting on http://account:%d/", cfg.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	serviceURL := fmt.Sprintf("http://account:%d/graphql", cfg.App.Port)
	schemaURL := fmt.Sprintf("http://account:%d/graphql", cfg.App.Port)

	// Register with etcd
	endpoints := strings.Split(cfg.Etcd.Endpoints, ",")
	reg, err := etcd.NewEtcdRegistry(endpoints, cfg.Etcd.BasePath, cfg.App.Name, serviceURL, schemaURL)
	if err != nil {
		log.Fatalf("Failed to create etcd registry: %v", err)
	}

	if err := reg.Register(); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// Set up graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down server...")

	// Create a deadline context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	// Unregister from etcd
	if err := reg.Unregister(); err != nil {
		log.Printf("Failed to unregister service: %v", err)
	}

	log.Println("Server stopped gracefully")
}
