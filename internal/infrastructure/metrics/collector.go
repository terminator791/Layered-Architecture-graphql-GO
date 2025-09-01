package metrics

import (
	"context"
	"sync"
	"time"
)

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

// Metric represents a single metric data point
type Metric struct {
	Name      string            `json:"name"`
	Type      MetricType        `json:"type"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// MetricsCollector provides metrics collection functionality
type MetricsCollector struct {
	mu      sync.RWMutex
	metrics map[string]*Metric
	
	// Application metrics
	requestCount      map[string]int64
	responseTime      map[string][]time.Duration
	activeConnections int64
	messagesSent      int64
	messagesReceived  int64
	
	// Business metrics
	totalUsers      int64
	totalRooms      int64
	totalMessages   int64
	activeUsers     int64
	threadsCreated  int64
	reactionsAdded  int64
	
	// System metrics
	startTime       time.Time
	lastHealthCheck time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics:           make(map[string]*Metric),
		requestCount:      make(map[string]int64),
		responseTime:      make(map[string][]time.Duration),
		activeConnections: 0,
		messagesSent:      0,
		messagesReceived:  0,
		totalUsers:        0,
		totalRooms:        0,
		totalMessages:     0,
		activeUsers:       0,
		threadsCreated:    0,
		reactionsAdded:    0,
		startTime:         time.Now(),
		lastHealthCheck:   time.Now(),
	}
}

// Counter operations

// IncrementCounter increments a counter metric
func (m *MetricsCollector) IncrementCounter(name string, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := m.buildMetricKey(name, labels)
	if metric, exists := m.metrics[key]; exists {
		metric.Value++
		metric.Timestamp = time.Now()
	} else {
		m.metrics[key] = &Metric{
			Name:      name,
			Type:      MetricTypeCounter,
			Value:     1,
			Labels:    labels,
			Timestamp: time.Now(),
		}
	}
}

// AddToCounter adds a value to a counter metric
func (m *MetricsCollector) AddToCounter(name string, value float64, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := m.buildMetricKey(name, labels)
	if metric, exists := m.metrics[key]; exists {
		metric.Value += value
		metric.Timestamp = time.Now()
	} else {
		m.metrics[key] = &Metric{
			Name:      name,
			Type:      MetricTypeCounter,
			Value:     value,
			Labels:    labels,
			Timestamp: time.Now(),
		}
	}
}

// Gauge operations

// SetGauge sets a gauge metric value
func (m *MetricsCollector) SetGauge(name string, value float64, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := m.buildMetricKey(name, labels)
	m.metrics[key] = &Metric{
		Name:      name,
		Type:      MetricTypeGauge,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}
}

// Histogram operations

// RecordDuration records a duration for histogram metrics
func (m *MetricsCollector) RecordDuration(name string, duration time.Duration, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := m.buildMetricKey(name, labels)
	
	// For simplicity, we'll store the average duration
	// In a real implementation, you'd want proper histogram buckets
	if metric, exists := m.metrics[key]; exists {
		// Calculate new average
		count := metric.Labels["count"]
		if count == "" {
			count = "1"
		}
		
		// This is a simplified histogram implementation
		metric.Value = (metric.Value + duration.Seconds()) / 2
		metric.Timestamp = time.Now()
	} else {
		if labels == nil {
			labels = make(map[string]string)
		}
		labels["count"] = "1"
		
		m.metrics[key] = &Metric{
			Name:      name,
			Type:      MetricTypeHistogram,
			Value:     duration.Seconds(),
			Labels:    labels,
			Timestamp: time.Now(),
		}
	}
}

// Application-specific metrics

// RecordRequest records an HTTP request
func (m *MetricsCollector) RecordRequest(method, endpoint string, duration time.Duration, statusCode int) {
	labels := map[string]string{
		"method":      method,
		"endpoint":    endpoint,
		"status_code": string(rune(statusCode)),
	}
	
	// Increment request count
	m.IncrementCounter("http_requests_total", labels)
	
	// Record response time
	m.RecordDuration("http_request_duration_seconds", duration, labels)
	
	// Update internal tracking
	m.mu.Lock()
	key := method + " " + endpoint
	m.requestCount[key]++
	m.responseTime[key] = append(m.responseTime[key], duration)
	
	// Keep only last 1000 response times for memory efficiency
	if len(m.responseTime[key]) > 1000 {
		m.responseTime[key] = m.responseTime[key][len(m.responseTime[key])-1000:]
	}
	m.mu.Unlock()
}

// RecordGraphQLOperation records a GraphQL operation
func (m *MetricsCollector) RecordGraphQLOperation(operationType, operationName string, duration time.Duration, hasErrors bool) {
	labels := map[string]string{
		"operation_type": operationType,
		"operation_name": operationName,
	}
	
	if hasErrors {
		labels["status"] = "error"
	} else {
		labels["status"] = "success"
	}
	
	m.IncrementCounter("graphql_operations_total", labels)
	m.RecordDuration("graphql_operation_duration_seconds", duration, labels)
}

// Business metrics

// RecordMessageSent records a message being sent
func (m *MetricsCollector) RecordMessageSent(roomType, messageType string) {
	labels := map[string]string{
		"room_type":    roomType,
		"message_type": messageType,
	}
	
	m.IncrementCounter("messages_sent_total", labels)
	
	m.mu.Lock()
	m.messagesSent++
	m.mu.Unlock()
}

// RecordThreadCreated records a thread being created
func (m *MetricsCollector) RecordThreadCreated(roomType string) {
	labels := map[string]string{
		"room_type": roomType,
	}
	
	m.IncrementCounter("threads_created_total", labels)
	
	m.mu.Lock()
	m.threadsCreated++
	m.mu.Unlock()
}

// RecordReactionAdded records a reaction being added
func (m *MetricsCollector) RecordReactionAdded(emoji string) {
	labels := map[string]string{
		"emoji": emoji,
	}
	
	m.IncrementCounter("reactions_added_total", labels)
	
	m.mu.Lock()
	m.reactionsAdded++
	m.mu.Unlock()
}

// RecordUserRegistration records a new user registration
func (m *MetricsCollector) RecordUserRegistration() {
	m.IncrementCounter("users_registered_total", nil)
	
	m.mu.Lock()
	m.totalUsers++
	m.mu.Unlock()
}

// RecordRoomCreated records a room being created
func (m *MetricsCollector) RecordRoomCreated(roomType string) {
	labels := map[string]string{
		"room_type": roomType,
	}
	
	m.IncrementCounter("rooms_created_total", labels)
	
	m.mu.Lock()
	m.totalRooms++
	m.mu.Unlock()
}

// Connection metrics

// RecordConnectionOpened records a WebSocket connection being opened
func (m *MetricsCollector) RecordConnectionOpened() {
	m.mu.Lock()
	m.activeConnections++
	m.mu.Unlock()
	
	m.SetGauge("websocket_connections_active", float64(m.activeConnections), nil)
	m.IncrementCounter("websocket_connections_total", map[string]string{"type": "opened"})
}

// RecordConnectionClosed records a WebSocket connection being closed
func (m *MetricsCollector) RecordConnectionClosed() {
	m.mu.Lock()
	m.activeConnections--
	if m.activeConnections < 0 {
		m.activeConnections = 0
	}
	m.mu.Unlock()
	
	m.SetGauge("websocket_connections_active", float64(m.activeConnections), nil)
	m.IncrementCounter("websocket_connections_total", map[string]string{"type": "closed"})
}

// System metrics

// UpdateSystemMetrics updates system-level metrics
func (m *MetricsCollector) UpdateSystemMetrics(ctx context.Context) {
	m.SetGauge("application_uptime_seconds", time.Since(m.startTime).Seconds(), nil)
	m.SetGauge("last_health_check_seconds", time.Since(m.lastHealthCheck).Seconds(), nil)
	
	// Record current totals as gauges
	m.mu.RLock()
	m.SetGauge("total_users", float64(m.totalUsers), nil)
	m.SetGauge("total_rooms", float64(m.totalRooms), nil)
	m.SetGauge("total_messages", float64(m.totalMessages), nil)
	m.SetGauge("active_connections", float64(m.activeConnections), nil)
	m.mu.RUnlock()
}

// Metric retrieval

// GetAllMetrics returns all collected metrics
func (m *MetricsCollector) GetAllMetrics() map[string]*Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Create a copy to avoid concurrent access issues
	result := make(map[string]*Metric)
	for key, metric := range m.metrics {
		result[key] = &Metric{
			Name:      metric.Name,
			Type:      metric.Type,
			Value:     metric.Value,
			Labels:    metric.Labels,
			Timestamp: metric.Timestamp,
		}
	}
	
	return result
}

// GetMetric returns a specific metric
func (m *MetricsCollector) GetMetric(name string, labels map[string]string) *Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	key := m.buildMetricKey(name, labels)
	if metric, exists := m.metrics[key]; exists {
		return &Metric{
			Name:      metric.Name,
			Type:      metric.Type,
			Value:     metric.Value,
			Labels:    metric.Labels,
			Timestamp: metric.Timestamp,
		}
	}
	
	return nil
}

// GetSummaryStats returns summary statistics
func (m *MetricsCollector) GetSummaryStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return map[string]interface{}{
		"application": map[string]interface{}{
			"uptime_seconds":      time.Since(m.startTime).Seconds(),
			"active_connections":  m.activeConnections,
			"total_requests":      m.getTotalRequests(),
			"average_response_ms": m.getAverageResponseTime(),
		},
		"business": map[string]interface{}{
			"total_users":     m.totalUsers,
			"total_rooms":     m.totalRooms,
			"total_messages":  m.totalMessages,
			"threads_created": m.threadsCreated,
			"reactions_added": m.reactionsAdded,
		},
	}
}

// Helper functions

// buildMetricKey creates a unique key for a metric
func (m *MetricsCollector) buildMetricKey(name string, labels map[string]string) string {
	key := name
	if labels != nil {
		for k, v := range labels {
			key += "|" + k + "=" + v
		}
	}
	return key
}

// getTotalRequests returns the total number of requests
func (m *MetricsCollector) getTotalRequests() int64 {
	var total int64
	for _, count := range m.requestCount {
		total += count
	}
	return total
}

// getAverageResponseTime returns the average response time in milliseconds
func (m *MetricsCollector) getAverageResponseTime() float64 {
	var totalDuration time.Duration
	var count int
	
	for _, durations := range m.responseTime {
		for _, duration := range durations {
			totalDuration += duration
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	
	return float64(totalDuration.Nanoseconds()) / float64(count) / 1000000 // Convert to milliseconds
}

// RecordHealthCheck updates the last health check timestamp
func (m *MetricsCollector) RecordHealthCheck() {
	m.mu.Lock()
	m.lastHealthCheck = time.Now()
	m.mu.Unlock()
}