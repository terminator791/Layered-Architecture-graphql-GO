# 📊 Monitoring & Observability Guide

A comprehensive guide to monitoring, health checks, and observability features in the real-time chat application. This demonstrates production-ready monitoring practices for Go applications.

## 📋 Table of Contents

1. [Overview](#overview)
2. [Health Checks](#health-checks)
3. [Metrics Collection](#metrics-collection)
4. [Prometheus Integration](#prometheus-integration)
5. [Performance Monitoring](#performance-monitoring)
6. [Alerting Strategy](#alerting-strategy)
7. [Deployment Integration](#deployment-integration)

## 🎯 Overview

The monitoring system provides comprehensive observability into:

- **Application Health**: Service availability and component status
- **Performance Metrics**: Response times, throughput, and resource usage
- **Business Metrics**: User activity, message volume, and feature usage
- **System Metrics**: Connection pools, cache performance, and infrastructure health

## 🏥 Health Checks

### Health Check Types

The application provides three types of health checks:

#### 1. Comprehensive Health Check (`/health`)

Provides detailed health information for all components:

```http
GET /health
```

**Response Example**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "uptime": "2h30m45s",
  "components": {
    "database": {
      "status": "healthy",
      "message": "Database is healthy",
      "lastChecked": "2024-01-15T10:30:00Z",
      "responseTime": "12ms"
    },
    "redis": {
      "status": "healthy",
      "message": "Redis is healthy",
      "lastChecked": "2024-01-15T10:30:00Z",
      "responseTime": "3ms"
    },
    "application": {
      "status": "healthy",
      "message": "Application running for 2h30m45s",
      "lastChecked": "2024-01-15T10:30:00Z",
      "responseTime": "1ms"
    }
  }
}
```

#### 2. Readiness Check (`/ready`)

Quick readiness check for Kubernetes:

```http
GET /ready
```

**Response**: `200 OK` with "Ready" or `503 Service Unavailable` with error message

#### 3. Liveness Check (`/live`)

Simple liveness check for Kubernetes:

```http
GET /live
```

**Response Example**:
```json
{
  "status": "alive",
  "timestamp": "2024-01-15T10:30:00Z",
  "uptime": "2h30m45s"
}
```

### Health Check Implementation

```go
// Create health checker
healthChecker := health.NewHealthChecker(db, redisClient, "1.0.0")

// Register health endpoints
http.HandleFunc("/health", healthChecker.HTTPHandler())
http.HandleFunc("/ready", healthChecker.ReadinessHandler())
http.HandleFunc("/live", healthChecker.LivenessHandler())
```

### Component Health Checks

#### Database Health
- **Connectivity**: Ping test with timeout
- **Query Execution**: Simple SELECT query
- **Connection Pool**: Monitor connection usage
- **Performance**: Response time thresholds

#### Redis Health
- **Connectivity**: Ping test with timeout
- **Operations**: Set/Get test with data integrity check
- **Performance**: Response time monitoring
- **Cleanup**: Automatic test data removal

#### Application Health
- **Uptime**: Application startup time tracking
- **State**: Basic application status checks
- **Dependencies**: Critical service availability

### Health Status Levels

- **Healthy**: All systems operational
- **Degraded**: Some issues but service available
- **Unhealthy**: Critical issues, service unavailable

## 📈 Metrics Collection

### Metric Types

The system collects three types of metrics:

#### Counters
Monotonically increasing values:
```go
// Increment message count
metrics.IncrementCounter("messages_sent_total", map[string]string{
    "room_type": "public",
    "message_type": "text",
})
```

#### Gauges
Point-in-time values:
```go
// Set active connections
metrics.SetGauge("websocket_connections_active", 150, nil)
```

#### Histograms
Duration measurements:
```go
// Record response time
metrics.RecordDuration("http_request_duration_seconds", duration, map[string]string{
    "method": "POST",
    "endpoint": "/graphql",
})
```

### Application Metrics

#### HTTP Request Metrics
- **Request Count**: Total requests by method, endpoint, status
- **Response Time**: Request duration histograms
- **Error Rate**: Failed request percentage

```go
// Record HTTP request
metrics.RecordRequest("POST", "/graphql", duration, 200)
```

#### GraphQL Operation Metrics
- **Operation Count**: Operations by type and name
- **Operation Duration**: Execution time tracking
- **Error Tracking**: Operation success/failure rates

```go
// Record GraphQL operation
metrics.RecordGraphQLOperation("mutation", "sendMessage", duration, false)
```

#### WebSocket Connection Metrics
- **Active Connections**: Current WebSocket connections
- **Connection Events**: Opens and closes
- **Message Volume**: Real-time message throughput

```go
// Track WebSocket connections
metrics.RecordConnectionOpened()
metrics.RecordConnectionClosed()
```

### Business Metrics

#### User Activity
- **User Registrations**: New user signups
- **Active Users**: Current online users
- **User Actions**: Login, logout, status changes

```go
// Track user registration
metrics.RecordUserRegistration()
```

#### Messaging Activity
- **Messages Sent**: Total message volume by type and room
- **Threads Created**: Thread creation activity
- **Reactions Added**: User engagement metrics

```go
// Track message activity
metrics.RecordMessageSent("public", "text")
metrics.RecordThreadCreated("private")
metrics.RecordReactionAdded("👍")
```

#### Room Activity
- **Room Creation**: New room creation by type
- **Room Membership**: Join/leave events
- **Room Usage**: Active rooms and member counts

```go
// Track room creation
metrics.RecordRoomCreated("public")
```

### System Metrics

#### Application Metrics
- **Uptime**: Application runtime
- **Memory Usage**: Memory consumption tracking
- **CPU Usage**: Processing load monitoring

#### Database Metrics
- **Connection Pool**: Active/idle connections
- **Query Performance**: Slow query detection
- **Transaction Volume**: Database operation rates

#### Cache Metrics
- **Hit/Miss Ratio**: Cache effectiveness
- **Eviction Rate**: Cache memory pressure
- **Response Time**: Cache performance

## 📊 Prometheus Integration

### Metrics Endpoint

The application provides Prometheus-compatible metrics:

```http
GET /metrics/prometheus
```

**Response Format**:
```
# HELP chat_app_metrics Chat application metrics
# TYPE chat_app_metrics counter
http_requests_total{method="POST",endpoint="/graphql",status_code="200"} 1234 1642234567000
websocket_connections_active 150 1642234567000
messages_sent_total{room_type="public",message_type="text"} 5678 1642234567000
```

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'chat-app'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics/prometheus
    scrape_interval: 10s
```

### Common Prometheus Queries

#### Request Rate
```promql
# Requests per second
rate(http_requests_total[5m])

# Request rate by endpoint
sum(rate(http_requests_total[5m])) by (endpoint)
```

#### Error Rate
```promql
# Error percentage
(
  sum(rate(http_requests_total{status_code=~"5.."}[5m])) 
  / 
  sum(rate(http_requests_total[5m]))
) * 100
```

#### Response Time
```promql
# Average response time
rate(http_request_duration_seconds_sum[5m]) 
/ 
rate(http_request_duration_seconds_count[5m])

# 95th percentile response time
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

#### Business Metrics
```promql
# Messages per minute
rate(messages_sent_total[1m]) * 60

# Active users
users_active

# Thread creation rate
rate(threads_created_total[5m])
```

## 🔍 Performance Monitoring

### Response Time Monitoring

Track response times across all endpoints:

```go
// Middleware for automatic response time tracking
func MetricsMiddleware(metrics *metrics.MetricsCollector) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        metrics.RecordRequest(
            c.Request.Method,
            c.FullPath(),
            duration,
            c.Writer.Status(),
        )
    }
}
```

### Database Performance

Monitor database query performance:

```go
// Database query wrapper with metrics
func (r *repository) QueryWithMetrics(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    start := time.Now()
    
    rows, err := r.db.QueryContext(ctx, query, args...)
    
    duration := time.Since(start)
    r.metrics.RecordDuration("database_query_duration_seconds", duration, map[string]string{
        "operation": "select",
    })
    
    return rows, err
}
```

### Memory and CPU Monitoring

Track resource usage:

```go
// Update system metrics periodically
func (m *MetricsCollector) UpdateSystemMetrics(ctx context.Context) {
    // Runtime metrics
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    m.SetGauge("memory_alloc_bytes", float64(memStats.Alloc), nil)
    m.SetGauge("memory_sys_bytes", float64(memStats.Sys), nil)
    m.SetGauge("goroutines_total", float64(runtime.NumGoroutine()), nil)
}
```

## 🚨 Alerting Strategy

### Critical Alerts (Immediate Response)

#### Service Availability
```yaml
# Prometheus Alert Rules
groups:
  - name: chat-app-critical
    rules:
      - alert: ServiceDown
        expr: up{job="chat-app"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Chat application is down"
          description: "The chat application has been down for more than 1 minute"

      - alert: HighErrorRate
        expr: (sum(rate(http_requests_total{status_code=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))) > 0.1
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} for more than 2 minutes"
```

#### Database Issues
```yaml
      - alert: DatabaseDown
        expr: database_status != 1
        for: 30s
        labels:
          severity: critical
        annotations:
          summary: "Database connection failed"

      - alert: SlowDatabaseQueries
        expr: database_query_duration_seconds > 5
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Database queries are slow"
```

### Warning Alerts (Monitor Closely)

#### Performance Degradation
```yaml
      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response time"

      - alert: HighMemoryUsage
        expr: memory_alloc_bytes > 1e9  # 1GB
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
```

#### Business Metrics
```yaml
      - alert: LowUserActivity
        expr: rate(messages_sent_total[1h]) < 10
        for: 30m
        labels:
          severity: warning
        annotations:
          summary: "Low user activity detected"
```

### Notification Channels

#### Slack Integration
```yaml
# alertmanager.yml
route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
  - name: 'web.hook'
    slack_configs:
      - api_url: 'YOUR_SLACK_WEBHOOK_URL'
        channel: '#alerts'
        title: 'Chat App Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}\n{{ .Annotations.description }}{{ end }}'
```

## 🚀 Deployment Integration

### Kubernetes Integration

#### Health Check Probes
```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: chat-app
spec:
  template:
    spec:
      containers:
      - name: chat-app
        image: chat-app:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
```

#### Service Monitor
```yaml
# servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: chat-app
spec:
  selector:
    matchLabels:
      app: chat-app
  endpoints:
  - port: http
    path: /metrics/prometheus
    interval: 15s
```

### Docker Compose Integration

```yaml
# docker-compose.yml
version: '3.8'
services:
  chat-app:
    build: .
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    environment:
      - METRICS_ENABLED=true
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana
```

### Grafana Dashboard

Create dashboards for monitoring:

```json
{
  "dashboard": {
    "title": "Chat Application Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "Active WebSocket Connections",
        "type": "stat",
        "targets": [
          {
            "expr": "websocket_connections_active",
            "legendFormat": "Active Connections"
          }
        ]
      },
      {
        "title": "Messages Sent",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(messages_sent_total[1m]) * 60",
            "legendFormat": "Messages per minute"
          }
        ]
      }
    ]
  }
}
```

This comprehensive monitoring and observability system provides complete visibility into the application's health, performance, and business metrics, enabling proactive monitoring and rapid issue resolution.