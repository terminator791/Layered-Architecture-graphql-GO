package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth represents the health of an individual component
type ComponentHealth struct {
	Status      HealthStatus `json:"status"`
	Message     string       `json:"message,omitempty"`
	LastChecked time.Time    `json:"lastChecked"`
	ResponseTime string      `json:"responseTime,omitempty"`
}

// HealthCheck represents the overall health check response
type HealthCheck struct {
	Status     HealthStatus               `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Version    string                     `json:"version"`
	Uptime     string                     `json:"uptime"`
	Components map[string]ComponentHealth `json:"components"`
}

// HealthChecker provides health checking functionality
type HealthChecker struct {
	db          *sqlx.DB
	redis       *redis.Client
	version     string
	startTime   time.Time
	checkTimeout time.Duration
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *sqlx.DB, redis *redis.Client, version string) *HealthChecker {
	return &HealthChecker{
		db:           db,
		redis:        redis,
		version:      version,
		startTime:    time.Now(),
		checkTimeout: 5 * time.Second,
	}
}

// CheckHealth performs a comprehensive health check
func (h *HealthChecker) CheckHealth(ctx context.Context) *HealthCheck {
	components := make(map[string]ComponentHealth)
	
	// Check database health
	dbHealth := h.checkDatabaseHealth(ctx)
	components["database"] = dbHealth
	
	// Check Redis health
	redisHealth := h.checkRedisHealth(ctx)
	components["redis"] = redisHealth
	
	// Check application health
	appHealth := h.checkApplicationHealth(ctx)
	components["application"] = appHealth
	
	// Determine overall status
	overallStatus := h.determineOverallStatus(components)
	
	return &HealthCheck{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Version:    h.version,
		Uptime:     time.Since(h.startTime).String(),
		Components: components,
	}
}

// checkDatabaseHealth checks PostgreSQL database connectivity and performance
func (h *HealthChecker) checkDatabaseHealth(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	// Create context with timeout
	dbCtx, cancel := context.WithTimeout(ctx, h.checkTimeout)
	defer cancel()
	
	// Test basic connectivity
	if err := h.db.PingContext(dbCtx); err != nil {
		return ComponentHealth{
			Status:       HealthStatusUnhealthy,
			Message:      fmt.Sprintf("Database ping failed: %v", err),
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start).String(),
		}
	}
	
	// Test query execution
	var result int
	query := "SELECT 1"
	if err := h.db.QueryRowxContext(dbCtx, query).Scan(&result); err != nil {
		return ComponentHealth{
			Status:       HealthStatusUnhealthy,
			Message:      fmt.Sprintf("Database query failed: %v", err),
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start).String(),
		}
	}
	
	// Check connection pool stats
	stats := h.db.Stats()
	if float64(stats.OpenConnections) > float64(stats.MaxOpenConnections)*0.9 {
		return ComponentHealth{
			Status:       HealthStatusDegraded,
			Message:      fmt.Sprintf("High connection usage: %d/%d", stats.OpenConnections, stats.MaxOpenConnections),
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start).String(),
		}
	}
	
	responseTime := time.Since(start)
	if responseTime > time.Second {
		return ComponentHealth{
			Status:       HealthStatusDegraded,
			Message:      "Slow database response",
			LastChecked:  time.Now(),
			ResponseTime: responseTime.String(),
		}
	}
	
	return ComponentHealth{
		Status:       HealthStatusHealthy,
		Message:      "Database is healthy",
		LastChecked:  time.Now(),
		ResponseTime: responseTime.String(),
	}
}

// checkRedisHealth checks Redis connectivity and performance
func (h *HealthChecker) checkRedisHealth(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	// Create context with timeout
	redisCtx, cancel := context.WithTimeout(ctx, h.checkTimeout)
	defer cancel()
	
	// Test basic connectivity
	if err := h.redis.Ping(redisCtx).Err(); err != nil {
		return ComponentHealth{
			Status:       HealthStatusUnhealthy,
			Message:      fmt.Sprintf("Redis ping failed: %v", err),
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start).String(),
		}
	}
	
	// Test set/get operation
	testKey := "health_check_test"
	testValue := fmt.Sprintf("test_%d", time.Now().Unix())
	
	if err := h.redis.Set(redisCtx, testKey, testValue, time.Minute).Err(); err != nil {
		return ComponentHealth{
			Status:       HealthStatusUnhealthy,
			Message:      fmt.Sprintf("Redis set failed: %v", err),
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start).String(),
		}
	}
	
	retrievedValue, err := h.redis.Get(redisCtx, testKey).Result()
	if err != nil {
		return ComponentHealth{
			Status:       HealthStatusUnhealthy,
			Message:      fmt.Sprintf("Redis get failed: %v", err),
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start).String(),
		}
	}
	
	if retrievedValue != testValue {
		return ComponentHealth{
			Status:       HealthStatusUnhealthy,
			Message:      "Redis data integrity check failed",
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start).String(),
		}
	}
	
	// Clean up test key
	h.redis.Del(redisCtx, testKey)
	
	responseTime := time.Since(start)
	if responseTime > 500*time.Millisecond {
		return ComponentHealth{
			Status:       HealthStatusDegraded,
			Message:      "Slow Redis response",
			LastChecked:  time.Now(),
			ResponseTime: responseTime.String(),
		}
	}
	
	return ComponentHealth{
		Status:       HealthStatusHealthy,
		Message:      "Redis is healthy",
		LastChecked:  time.Now(),
		ResponseTime: responseTime.String(),
	}
}

// checkApplicationHealth checks application-level health metrics
func (h *HealthChecker) checkApplicationHealth(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	// Check if we can perform basic operations
	// This could include checking critical business logic, cache coherence, etc.
	
	// For now, we'll just check uptime and basic application state
	uptime := time.Since(h.startTime)
	
	if uptime < time.Minute {
		return ComponentHealth{
			Status:       HealthStatusDegraded,
			Message:      "Application recently started",
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start).String(),
		}
	}
	
	return ComponentHealth{
		Status:       HealthStatusHealthy,
		Message:      fmt.Sprintf("Application running for %s", uptime.String()),
		LastChecked:  time.Now(),
		ResponseTime: time.Since(start).String(),
	}
}

// determineOverallStatus determines the overall health status based on component health
func (h *HealthChecker) determineOverallStatus(components map[string]ComponentHealth) HealthStatus {
	hasUnhealthy := false
	hasDegraded := false
	
	for _, component := range components {
		switch component.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}
	
	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}
	
	return HealthStatusHealthy
}

// HTTPHandler returns an HTTP handler for health checks
func (h *HealthChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Perform health check
		health := h.CheckHealth(ctx)
		
		// Set appropriate HTTP status code
		var statusCode int
		switch health.Status {
		case HealthStatusHealthy:
			statusCode = http.StatusOK
		case HealthStatusDegraded:
			statusCode = http.StatusOK // Still OK, but with warnings
		case HealthStatusUnhealthy:
			statusCode = http.StatusServiceUnavailable
		default:
			statusCode = http.StatusInternalServerError
		}
		
		// Set headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.WriteHeader(statusCode)
		
		// Encode and send response
		if err := json.NewEncoder(w).Encode(health); err != nil {
			http.Error(w, "Failed to encode health check response", http.StatusInternalServerError)
			return
		}
	}
}

// ReadinessHandler returns a simpler readiness check for Kubernetes
func (h *HealthChecker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Quick database connectivity check
		dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		
		if err := h.db.PingContext(dbCtx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Database not ready"))
			return
		}
		
		// Quick Redis connectivity check
		redisCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		
		if err := h.redis.Ping(redisCtx).Err(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Redis not ready"))
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	}
}

// LivenessHandler returns a simple liveness check for Kubernetes
func (h *HealthChecker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple check that the application is running
		// This should only fail if the application is completely broken
		
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"status":    "alive",
			"timestamp": time.Now(),
			"uptime":    time.Since(h.startTime).String(),
		}
		
		json.NewEncoder(w).Encode(response)
	}
}