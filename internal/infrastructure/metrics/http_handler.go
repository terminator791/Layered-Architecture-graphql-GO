package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// HTTPHandler provides HTTP endpoints for metrics
type HTTPHandler struct {
	collector *MetricsCollector
}

// NewHTTPHandler creates a new metrics HTTP handler
func NewHTTPHandler(collector *MetricsCollector) *HTTPHandler {
	return &HTTPHandler{
		collector: collector,
	}
}

// MetricsHandler returns all metrics in JSON format
func (h *HTTPHandler) MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := h.collector.GetAllMetrics()
		
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		
		response := map[string]interface{}{
			"metrics":   metrics,
			"timestamp": h.collector.startTime,
		}
		
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
			return
		}
	}
}

// PrometheusHandler returns metrics in Prometheus format
func (h *HTTPHandler) PrometheusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := h.collector.GetAllMetrics()
		
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		
		var output strings.Builder
		
		// Add metadata comments
		output.WriteString("# HELP chat_app_metrics Chat application metrics\n")
		output.WriteString("# TYPE chat_app_metrics counter\n")
		
		for _, metric := range metrics {
			// Convert metric to Prometheus format
			prometheusName := sanitizeMetricName(metric.Name)
			
			// Write metric with labels
			if len(metric.Labels) > 0 {
				output.WriteString(fmt.Sprintf("%s{", prometheusName))
				labelPairs := make([]string, 0, len(metric.Labels))
				for k, v := range metric.Labels {
					labelPairs = append(labelPairs, fmt.Sprintf(`%s="%s"`, sanitizeLabelName(k), sanitizeLabelValue(v)))
				}
				output.WriteString(strings.Join(labelPairs, ","))
				output.WriteString(fmt.Sprintf("} %g %d\n", metric.Value, metric.Timestamp.Unix()*1000))
			} else {
				output.WriteString(fmt.Sprintf("%s %g %d\n", prometheusName, metric.Value, metric.Timestamp.Unix()*1000))
			}
		}
		
		w.Write([]byte(output.String()))
	}
}

// SummaryHandler returns summary statistics
func (h *HTTPHandler) SummaryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		summary := h.collector.GetSummaryStats()
		
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		
		if err := json.NewEncoder(w).Encode(summary); err != nil {
			http.Error(w, "Failed to encode summary", http.StatusInternalServerError)
			return
		}
	}
}

// RegisterRoutes registers metrics routes with an HTTP mux
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/metrics", h.MetricsHandler())
	mux.HandleFunc("/metrics/prometheus", h.PrometheusHandler())
	mux.HandleFunc("/metrics/summary", h.SummaryHandler())
}

// Helper functions for Prometheus format

// sanitizeMetricName converts metric names to Prometheus format
func sanitizeMetricName(name string) string {
	// Replace invalid characters with underscores
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, " ", "_")
	
	// Ensure it starts with a letter or underscore
	if len(name) > 0 && (name[0] >= '0' && name[0] <= '9') {
		name = "_" + name
	}
	
	return strings.ToLower(name)
}

// sanitizeLabelName converts label names to Prometheus format
func sanitizeLabelName(name string) string {
	return sanitizeMetricName(name)
}

// sanitizeLabelValue escapes label values for Prometheus format
func sanitizeLabelValue(value string) string {
	// Escape backslashes and quotes
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\t", "\\t")
	return value
}