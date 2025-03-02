package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gianglt2198/graphql-gateway-go/gateway/internal/etcd"
	"github.com/gianglt2198/graphql-gateway-go/gateway/internal/gateway"
	"github.com/samber/lo"
	"github.com/wundergraph/graphql-go-tools/pkg/playground"
)

func main() {
	// Parse command line flags
	etcdEndpoints := flag.String("etcd-endpoints", "localhost:2379", "Comma-separated list of etcd endpoints")
	etcdBasePath := flag.String("etcd-base-path", "/services", "Base path for service discovery in etcd")
	listenAddr := flag.String("listen", ":8080", "Address to listen on")
	flag.Parse()

	// Create etcd client
	endpoints := strings.Split(lo.FromPtr(etcdEndpoints), ",")
	etcdClient, err := etcd.NewClient(endpoints, *etcdBasePath)
	if err != nil {
		log.Fatalf("Failed to create etcd client: %v", err)
	}
	defer etcdClient.Close()

	mux := http.NewServeMux()

	// Create gateway server
	server, err := gateway.NewServer(etcdClient, mux)
	if err != nil {
		log.Fatalf("Failed to create gateway server: %v", err)
	}

	graphqlEndpoint := "/graphql"
	playgroundURLPrefix := "/playground"
	// playgroundURL := ""

	p := playground.New(playground.Config{
		PathPrefix:                      "",
		PlaygroundPath:                  playgroundURLPrefix,
		GraphqlEndpointPath:             graphqlEndpoint,
		GraphQLSubscriptionEndpointPath: graphqlEndpoint,
	})

	handlers, err := p.Handlers()
	if err != nil {
		log.Fatal("configure handlers", err)
		return
	}

	for i := range handlers {
		mux.Handle(handlers[i].Path, handlers[i].Handler)
	}

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    *listenAddr,
		Handler: server,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting GraphQL Federation Gateway on %s", *listenAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Set up graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Create a deadline context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}
