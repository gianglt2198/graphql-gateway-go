package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gianglt2198/graphql-gateway-go/gateway/internal/etcd"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

type SchemaRegistry struct {
	etcdClient     *etcd.Client
	schemas        map[string]*ast.Schema
	sdls           map[string]string
	schemasMutex   sync.Mutex
	listeners      map[string][]chan struct{}
	listenersMutex sync.Mutex
}

func NewSchemaRegistry(etcdClient *etcd.Client) *SchemaRegistry {
	registry := &SchemaRegistry{
		etcdClient: etcdClient,
		schemas:    make(map[string]*ast.Schema),
		sdls:       make(map[string]string),
		listeners:  make(map[string][]chan struct{}),
	}

	// Initialize schemas for all services
	services := etcdClient.GetAllServices()
	for name, service := range services {
		registry.fetchAndRegisterSchema(name, service.SchemaURL)
	}

	// Register for service updates
	go registry.watchServiceChanges()

	return registry
}

func (r *SchemaRegistry) watchServiceChanges() {
	globalListener := "*"
	updates, unregister := r.etcdClient.RegisterServiceListener(globalListener)

	go func() {
		defer unregister()

		for service := range updates {
			if service == nil {
				continue
			}

			if service.Type == "remove" {
				// Service was deleted
				r.removeSchema(service.Name)
			} else {
				// Service was added
				r.fetchAndRegisterSchema(service.Name, service.SchemaURL)
			}

			// Notify listeners
			r.notifySchemaChanged(globalListener)
		}
	}()
}

func (r *SchemaRegistry) fetchAndRegisterSchema(serviceName, schemaURL string) {
	sdl, err := r.fetchSchemaSDL(schemaURL)
	if err != nil {
		fmt.Printf("Error fetching schema for servie %s: %v\n", serviceName, err)
		return
	}

	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name:  serviceName,
		Input: sdl,
	})
	if err != nil {
		fmt.Printf("Error fetching schema for servie %s: %v\n", serviceName, err)
		return
	}

	r.schemasMutex.Lock()
	r.schemas[serviceName] = schema
	r.sdls[serviceName] = sdl
	r.schemasMutex.Unlock()

	fmt.Printf("Registered schema for service: %s\n", serviceName)
}

// removeSchema removes a schema for a service
func (r *SchemaRegistry) removeSchema(serviceName string) {
	r.schemasMutex.Lock()
	delete(r.schemas, serviceName)
	delete(r.sdls, serviceName)
	r.schemasMutex.Unlock()

	fmt.Printf("Removed schema for service: %s\n", serviceName)
}

type (
	SchemaResult struct {
		Data struct {
			Service struct {
				SDL string `json:"sdl"`
			} `json:"_service"`
		} `json:"data"`
		Errors []interface{} `json:"errors"`
	}
)

func (r *SchemaRegistry) fetchSchemaSDL(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", err
	}

	// Create the GraphQL introspection query
	query := `{
	    _service {
	        sdl
	    }
	}`

	reqBody, _ := json.Marshal(map[string]interface{}{
		"query": query,
	})

	req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result SchemaResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("error decoding ", err)
		return "", err
	}

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("GrapQL Errors: %v", result.Errors...)
	}

	return result.Data.Service.SDL, nil
}

func (r *SchemaRegistry) GetSchema(serviceName string) (*ast.Schema, bool) {
	r.schemasMutex.Lock()
	defer r.schemasMutex.Unlock()

	schema, ok := r.schemas[serviceName]
	return schema, ok
}

func (r *SchemaRegistry) RegisterSchemaListener(serviceName string) (<-chan struct{}, func()) {
	r.listenersMutex.Lock()
	defer r.listenersMutex.Unlock()

	ch := make(chan struct{}, 1)

	if _, ok := r.listeners[serviceName]; !ok {
		r.listeners[serviceName] = []chan struct{}{}
	}

	r.listeners[serviceName] = append(r.listeners[serviceName], ch)

	unregister := func() {
		r.listenersMutex.Lock()
		defer r.listenersMutex.Unlock()

		listeners := r.listeners[serviceName]
		for i, listener := range listeners {
			if listener == ch {
				r.listeners[serviceName] = append(listeners[:i], listeners[:i+1]...)
				close(ch)
				break
			}
		}
	}

	return ch, unregister
}

// notifySchemaChanged notifies all listeners of schema changes
func (r *SchemaRegistry) notifySchemaChanged(serviceName string) {
	r.listenersMutex.Lock()
	defer r.listenersMutex.Unlock()

	// Notify service-specific listeners
	if listeners, ok := r.listeners[serviceName]; ok {
		for _, listener := range listeners {
			select {
			case listener <- struct{}{}:
				// Notification sent
				log.Println("Notification registry to listerners successfully")
			default:
				// Channel is full, skip
			}
		}
	}
}

func (r *SchemaRegistry) GetAllSchemas() map[string]*ast.Schema {
	r.schemasMutex.Lock()
	defer r.schemasMutex.Unlock()

	// Create a copy to avoid race conditions
	schemas := make(map[string]*ast.Schema, len(r.schemas))
	for k, v := range r.schemas {
		schemas[k] = v
	}

	return schemas
}

func (r *SchemaRegistry) GetAllSDLs() map[string]string {
	r.schemasMutex.Lock()
	defer r.schemasMutex.Unlock()

	// Create a copy to avoid race conditions
	sdls := make(map[string]string, len(r.sdls))
	for k, v := range r.sdls {
		sdls[k] = v
	}

	return sdls
}

func (r *SchemaRegistry) GetServiceEndpoints() map[string]string {
	r.schemasMutex.Lock()
	defer r.schemasMutex.Unlock()

	// Get service endpoints from etcd
	services := r.etcdClient.GetAllServices()

	endpoints := make(map[string]string)
	for name, service := range services {
		// Extract base URL from the schema URL
		// Assuming schemaURL is like "http://service:8080/graphql"
		schemaURL := service.SchemaURL
		// Remove "/graphql" if it exists to get the base URL
		// baseURL := strings.TrimSuffix(schemaURL, "/graphql")
		endpoints[name] = schemaURL
	}

	return endpoints
}
