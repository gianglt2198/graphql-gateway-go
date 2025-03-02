package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type ServiceInfo struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	SchemaURL string `json:"schema_url"`
	Type      string `json:"type"`
}

type Client struct {
	client    *clientv3.Client
	services  map[string]*ServiceInfo
	listeners map[string][]chan *ServiceInfo
	mu        sync.RWMutex
	basePath  string
}

func NewClient(enpoints []string, basePath string) (*Client, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   enpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	client := &Client{
		client:    etcdClient,
		services:  make(map[string]*ServiceInfo),
		listeners: make(map[string][]chan *ServiceInfo),
		basePath:  basePath,
	}

	if err := client.discoverService(); err != nil {
		return nil, err
	}

	go client.watchServices()

	return client, err
}

func (c *Client) discoverService() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.client.Get(ctx, c.basePath, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	for _, kv := range res.Kvs {
		var service ServiceInfo
		if err := json.Unmarshal(kv.Value, &service); err != nil {
			log.Printf("Error unmarshaling service: %v", err)
			return err
		}
		fmt.Println("service", service)

		c.mu.Lock()
		c.services[service.Name] = &service
		c.mu.Unlock()
	}

	return nil
}

func (c *Client) watchServices() {
	watcher := c.client.Watch(context.Background(), c.basePath, clientv3.WithPrefix())
	for resp := range watcher {
		for _, ev := range resp.Events {
			key := string(ev.Kv.Key)
			serviceName := key[len(c.basePath)+1:] // skip the trailing slash

			switch ev.Type {
			case clientv3.EventTypePut:
				var service ServiceInfo
				if err := json.Unmarshal(ev.Kv.Value, &service); err != nil {
					log.Printf("Error unmarshaling service: %v", err)
					continue
				}

				service.Type = "add"

				c.mu.Lock()
				c.services[serviceName] = &service
				c.notifyListeners(serviceName, &service)
				c.mu.Unlock()

				log.Printf("Service updated: %s - %v", serviceName, &service)
			case clientv3.EventTypeDelete:
				service, ok := c.services[serviceName]
				if !ok {
					log.Println("Not found services", serviceName)
					continue
				}

				service.Type = "remove"

				c.mu.Lock()
				delete(c.services, serviceName)
				c.notifyListeners(serviceName, service) // nil indicates deletion
				c.mu.Unlock()

				log.Printf("Service deleted: %s", serviceName)
			}
		}
	}
}

// notifyListeners notifies all listeners of service changes
func (c *Client) notifyListeners(serviceName string, info *ServiceInfo) {
	// notify global listeners
	globalListeners, ok := c.listeners["*"]
	if !ok {
		fmt.Println("no found global listeners")
		return
	}
	for _, listener := range globalListeners {
		select {
		case listener <- info:
			// Notification sent
			log.Println("Notificatin sent to listeners", serviceName)
		default:
			// Channel is full or closed, skip
		}
	}
}

func (c *Client) RegisterServiceListener(serviceName string) (<-chan *ServiceInfo, func()) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan *ServiceInfo, 10) // Buffer to advoid blocking

	if _, ok := c.listeners[serviceName]; !ok {
		c.listeners[serviceName] = []chan *ServiceInfo{}
	}

	c.listeners[serviceName] = append(c.listeners[serviceName], ch)

	// If service already exists send initial notification
	if service, ok := c.services[serviceName]; ok {
		go func() {
			ch <- service
		}()
	}

	unregister := func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		listeners := c.listeners[serviceName]
		for i, listener := range listeners {
			for listener == ch {
				c.listeners[serviceName] = append(c.listeners[serviceName], listeners[i+1:]...)
				close(ch)
				break
			}
		}
	}

	return ch, unregister
}

// Close closes the etcd client
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close all listener channels
	for _, listeners := range c.listeners {
		for _, listener := range listeners {
			close(listener)
		}
	}

	return c.client.Close()
}

// GetService returns service information by name
func (c *Client) GetService(name string) (*ServiceInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	service, ok := c.services[name]
	return service, ok
}

// GetAllServices returns all registered services
func (c *Client) GetAllServices() map[string]*ServiceInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create a copy to avoid race conditions
	services := make(map[string]*ServiceInfo, len(c.services))
	for k, v := range c.services {
		services[k] = v
	}

	return services
}
