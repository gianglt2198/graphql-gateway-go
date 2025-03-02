package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type ServiceInfo struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	SchemaURL string `json:"schema_url"`
}

type EtcdRegistry struct {
	client     *clientv3.Client
	serviceKey string
	info       ServiceInfo
	leaseID    clientv3.LeaseID
	keepAlive  <-chan *clientv3.LeaseKeepAliveResponse
	stopCh     chan struct{}
}

func NewEtcdRegistry(endpoints []string, basePath, serviceName, serviceURL, schemaURL string) (*EtcdRegistry, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to etcd: %w", err)
	}

	serviceKey := fmt.Sprintf("%s/%s", basePath, serviceName)

	registry := &EtcdRegistry{
		client:     client,
		serviceKey: serviceKey,
		info: ServiceInfo{
			Name:      serviceName,
			URL:       serviceURL,
			SchemaURL: schemaURL,
		},
		stopCh: make(chan struct{}),
	}

	return registry, nil
}

func (r *EtcdRegistry) Register() error {
	resp, err := r.client.Grant(context.Background(), 10)
	if err != nil {
		return fmt.Errorf("failed to create lease: %w", err)
	}

	r.leaseID = resp.ID

	data, err := json.Marshal(r.info)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	_, err = r.client.Put(context.Background(), r.serviceKey, string(data), clientv3.WithLease(r.leaseID))
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	keepAlive, err := r.client.KeepAlive(context.Background(), r.leaseID)
	if err != nil {
		return fmt.Errorf("failed to keep lease alive: %w", err)
	}
	r.keepAlive = keepAlive
	go r.handleKeepAlive()
	log.Printf("Registered service %s at %s", r.info.Name, r.info.URL)
	return nil
}

func (r *EtcdRegistry) handleKeepAlive() {
	for {
		select {
		case <-r.stopCh:
			return
		case _, ok := <-r.keepAlive:
			if !ok {
				log.Println("Lease keep alive channal closed, attemping to re-register")
				if err := r.Register(); err != nil {
					log.Printf("Failed to re-register service: %v", err)
				}
				return
			}
			// log.Printf("Lease keep alive TTL: %d", resp.TTL)
		}
	}
}

func (r *EtcdRegistry) Unregister() error {
	close(r.stopCh)

	_, err := r.client.Delete(context.Background(), r.serviceKey)
	if err != nil {
		return fmt.Errorf("failed to unregister service: %w", err)
	}

	_, err = r.client.Revoke(context.Background(), r.leaseID)
	if err != nil {
		return fmt.Errorf("failed to revoke lease: %w", err)
	}

	log.Printf("Unregistered service %s", r.info.Name)
	return r.client.Close()
}
