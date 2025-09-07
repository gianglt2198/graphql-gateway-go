package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheck represents a single health check
type HealthCheck struct {
	Name        string                                      `json:"name"`
	Description string                                      `json:"description"`
	CheckFunc   func(ctx context.Context) HealthCheckResult `json:"-"`
	Timeout     time.Duration                               `json:"timeout"`
	Interval    time.Duration                               `json:"interval"`
	Critical    bool                                        `json:"critical"`
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status    HealthStatus  `json:"status"`
	Message   string        `json:"message"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Details   interface{}   `json:"details,omitempty"`
}

// HealthChecker manages health checks for a service
type HealthChecker struct {
	serviceName string
	checks      map[string]*HealthCheck
	results     map[string]HealthCheckResult
	mu          sync.RWMutex
	logger      *logging.Logger
}

// HealthResponse represents the overall health response
type HealthResponse struct {
	Service   string                       `json:"service"`
	Status    HealthStatus                 `json:"status"`
	Timestamp time.Time                    `json:"timestamp"`
	Uptime    time.Duration                `json:"uptime"`
	Version   string                       `json:"version"`
	Checks    map[string]HealthCheckResult `json:"checks"`
	Summary   map[HealthStatus]int         `json:"summary"`
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(cfg *config.AppConfig, logger *logging.Logger) *HealthChecker {
	return &HealthChecker{
		serviceName: cfg.Name,
		checks:      make(map[string]*HealthCheck),
		results:     make(map[string]HealthCheckResult),
		logger:      logger,
	}
}

// RegisterCheck registers a new health check
func (hc *HealthChecker) RegisterCheck(check *HealthCheck) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.checks[check.Name] = check

	// Initialize result with unknown status
	hc.results[check.Name] = HealthCheckResult{
		Status:    HealthStatusUnknown,
		Message:   "Not yet checked",
		Timestamp: time.Now(),
		Duration:  0,
	}

	if hc.logger != nil {
		hc.logger.GetLogger().Info("Health check registered",
			zap.String("check_name", check.Name),
			zap.Bool("critical", check.Critical),
			zap.Duration("timeout", check.Timeout),
			zap.Duration("interval", check.Interval))
	}
}

// UnregisterCheck removes a health check
func (hc *HealthChecker) UnregisterCheck(name string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	delete(hc.checks, name)
	delete(hc.results, name)

	if hc.logger != nil {
		hc.logger.GetLogger().Info("Health check unregistered",
			zap.String("check_name", name))
	}
}

// RunCheck executes a specific health check
func (hc *HealthChecker) RunCheck(ctx context.Context, name string) HealthCheckResult {
	hc.mu.RLock()
	check, exists := hc.checks[name]
	hc.mu.RUnlock()

	if !exists {
		return HealthCheckResult{
			Status:    HealthStatusUnknown,
			Message:   fmt.Sprintf("Health check '%s' not found", name),
			Timestamp: time.Now(),
			Duration:  0,
		}
	}

	start := time.Now()

	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, check.Timeout)
	defer cancel()

	// Run the check
	result := check.CheckFunc(checkCtx)
	result.Timestamp = time.Now()
	result.Duration = time.Since(start)

	// Store the result
	hc.mu.Lock()
	hc.results[name] = result
	hc.mu.Unlock()

	// Log the result
	if hc.logger != nil {
		if result.Status != HealthStatusHealthy {
			hc.logger.GetLogger().Error("Health check completed",
				zap.String("check_name", name),
				zap.String("status", string(result.Status)),
				zap.String("message", result.Message),
				zap.Duration("duration", result.Duration))
		} else {
			hc.logger.GetLogger().Info("Health check completed",
				zap.String("check_name", name),
				zap.String("status", string(result.Status)),
				zap.String("message", result.Message),
				zap.Duration("duration", result.Duration))
		}
	}

	return result
}

// RunAllChecks executes all registered health checks
func (hc *HealthChecker) RunAllChecks(ctx context.Context) map[string]HealthCheckResult {
	hc.mu.RLock()
	checks := make(map[string]*HealthCheck)
	for name, check := range hc.checks {
		checks[name] = check
	}
	hc.mu.RUnlock()

	results := make(map[string]HealthCheckResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name := range checks {
		wg.Add(1)
		go func(checkName string) {
			defer wg.Done()
			result := hc.RunCheck(ctx, checkName)
			mu.Lock()
			results[checkName] = result
			mu.Unlock()
		}(name)
	}

	wg.Wait()
	return results
}

// GetHealth returns the overall health status
func (hc *HealthChecker) GetHealth(ctx context.Context) HealthResponse {
	results := hc.RunAllChecks(ctx)

	// Calculate overall status
	overallStatus := hc.calculateOverallStatus(results)

	// Calculate summary
	summary := make(map[HealthStatus]int)
	for _, result := range results {
		summary[result.Status]++
	}

	return HealthResponse{
		Service:   hc.serviceName,
		Status:    overallStatus,
		Timestamp: time.Now(),
		Uptime:    time.Since(time.Now()), // This should be set by the service
		Version:   "unknown",              // This should be set by the service
		Checks:    results,
		Summary:   summary,
	}
}

// calculateOverallStatus determines the overall health status based on individual checks
func (hc *HealthChecker) calculateOverallStatus(results map[string]HealthCheckResult) HealthStatus {
	if len(results) == 0 {
		return HealthStatusUnknown
	}

	hc.mu.RLock()
	defer hc.mu.RUnlock()

	hasCriticalFailure := false
	hasNonCriticalFailure := false
	hasUnknown := false

	for name, result := range results {
		check, exists := hc.checks[name]
		if !exists {
			continue
		}

		switch result.Status {
		case HealthStatusUnhealthy:
			if check.Critical {
				hasCriticalFailure = true
			} else {
				hasNonCriticalFailure = true
			}
		case HealthStatusDegraded:
			hasNonCriticalFailure = true
		case HealthStatusUnknown:
			hasUnknown = true
		}
	}

	if hasCriticalFailure {
		return HealthStatusUnhealthy
	}
	if hasNonCriticalFailure {
		return HealthStatusDegraded
	}
	if hasUnknown {
		return HealthStatusUnknown
	}

	return HealthStatusHealthy
}

// HTTPHandler returns an HTTP handler for health checks
func (hc *HealthChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Check if specific check is requested
		checkName := r.URL.Query().Get("check")
		if checkName != "" {
			result := hc.RunCheck(ctx, checkName)
			w.Header().Set("Content-Type", "application/json")

			statusCode := http.StatusOK
			if result.Status == HealthStatusUnhealthy {
				statusCode = http.StatusServiceUnavailable
			} else if result.Status == HealthStatusDegraded {
				statusCode = http.StatusPartialContent
			}

			w.WriteHeader(statusCode)
			_ = json.NewEncoder(w).Encode(result)
			return
		}

		// Return overall health
		health := hc.GetHealth(ctx)
		w.Header().Set("Content-Type", "application/json")

		statusCode := http.StatusOK
		if health.Status == HealthStatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		} else if health.Status == HealthStatusDegraded {
			statusCode = http.StatusPartialContent
		}

		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(health)
	}
}

// StartPeriodicChecks starts periodic execution of health checks
func (hc *HealthChecker) StartPeriodicChecks(ctx context.Context) {
	hc.mu.RLock()
	checks := make(map[string]*HealthCheck)
	for name, check := range hc.checks {
		if check.Interval > 0 {
			checks[name] = check
		}
	}
	hc.mu.RUnlock()

	for name, check := range checks {
		go hc.runPeriodicCheck(ctx, name, check.Interval)
	}
}

// runPeriodicCheck runs a health check periodically
func (hc *HealthChecker) runPeriodicCheck(ctx context.Context, name string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.RunCheck(ctx, name)
		}
	}
}

// Common health check functions

// DatabaseHealthCheck creates a health check for database connectivity
func DatabaseHealthCheck(name, description string, pingFunc func(ctx context.Context) error) *HealthCheck {
	return &HealthCheck{
		Name:        name,
		Description: description,
		Timeout:     5 * time.Second,
		Interval:    30 * time.Second,
		Critical:    true,
		CheckFunc: func(ctx context.Context) HealthCheckResult {
			err := pingFunc(ctx)
			if err != nil {
				return HealthCheckResult{
					Status:  HealthStatusUnhealthy,
					Message: fmt.Sprintf("Database connection failed: %v", err),
				}
			}
			return HealthCheckResult{
				Status:  HealthStatusHealthy,
				Message: "Database connection successful",
			}
		},
	}
}

// RedisHealthCheck creates a health check for Redis connectivity
func RedisHealthCheck(name, description string, pingFunc func(ctx context.Context) error) *HealthCheck {
	return &HealthCheck{
		Name:        name,
		Description: description,
		Timeout:     3 * time.Second,
		Interval:    30 * time.Second,
		Critical:    false,
		CheckFunc: func(ctx context.Context) HealthCheckResult {
			err := pingFunc(ctx)
			if err != nil {
				return HealthCheckResult{
					Status:  HealthStatusUnhealthy,
					Message: fmt.Sprintf("Redis connection failed: %v", err),
				}
			}
			return HealthCheckResult{
				Status:  HealthStatusHealthy,
				Message: "Redis connection successful",
			}
		},
	}
}

// NATSHealthCheck creates a health check for NATS connectivity
func NATSHealthCheck(name, description string, pingFunc func(ctx context.Context) error) *HealthCheck {
	return &HealthCheck{
		Name:        name,
		Description: description,
		Timeout:     3 * time.Second,
		Interval:    30 * time.Second,
		Critical:    false,
		CheckFunc: func(ctx context.Context) HealthCheckResult {
			err := pingFunc(ctx)
			if err != nil {
				return HealthCheckResult{
					Status:  HealthStatusUnhealthy,
					Message: fmt.Sprintf("NATS connection failed: %v", err),
				}
			}
			return HealthCheckResult{
				Status:  HealthStatusHealthy,
				Message: "NATS connection successful",
			}
		},
	}
}

// ETCDHealthCheck creates a health check for etcd connectivity
func ETCDHealthCheck(name, description string, pingFunc func(ctx context.Context) error) *HealthCheck {
	return &HealthCheck{
		Name:        name,
		Description: description,
		Timeout:     5 * time.Second,
		Interval:    30 * time.Second,
		Critical:    false,
		CheckFunc: func(ctx context.Context) HealthCheckResult {
			err := pingFunc(ctx)
			if err != nil {
				return HealthCheckResult{
					Status:  HealthStatusUnhealthy,
					Message: fmt.Sprintf("etcd connection failed: %v", err),
				}
			}
			return HealthCheckResult{
				Status:  HealthStatusHealthy,
				Message: "etcd connection successful",
			}
		},
	}
}

// ServiceHealthCheck creates a health check for external service connectivity
func ServiceHealthCheck(name, description, url string, httpClient *http.Client) *HealthCheck {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	return &HealthCheck{
		Name:        name,
		Description: description,
		Timeout:     10 * time.Second,
		Interval:    60 * time.Second,
		Critical:    false,
		CheckFunc: func(ctx context.Context) HealthCheckResult {
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return HealthCheckResult{
					Status:  HealthStatusUnhealthy,
					Message: fmt.Sprintf("Failed to create request: %v", err),
				}
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				return HealthCheckResult{
					Status:  HealthStatusUnhealthy,
					Message: fmt.Sprintf("Service request failed: %v", err),
				}
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return HealthCheckResult{
					Status:  HealthStatusHealthy,
					Message: fmt.Sprintf("Service responded with status %d", resp.StatusCode),
				}
			}

			return HealthCheckResult{
				Status:  HealthStatusUnhealthy,
				Message: fmt.Sprintf("Service responded with status %d", resp.StatusCode),
			}
		},
	}
}
