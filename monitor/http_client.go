package monitor

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// A HTTPClientHandler is a interface used by a Monitor to handle http client monitoring handler
type HTTPClientHandler interface {
	//	Init(namespace string)
	HandleRequest(r *http.Request)
	HandleResponse(r *http.Response)
	HandleLatency(r *http.Response, latency float64)
}

// RoundTripWrapper round trip wrapper
type RoundTripWrapper struct {
	hcHandler HTTPClientHandler
	transport http.RoundTripper
}

// RoundTrip executes a single HTTP transaction, returning
// a Response for the provided Request. It is used by HTTP client monitor.
func (m *RoundTripWrapper) RoundTrip(r *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := m.transport.RoundTrip(r)
	latency := (float64)(time.Since(start) / time.Millisecond)
	if err != nil {
		return nil, err
	}
	
	m.hcHandler.HandleRequest(r)

	m.hcHandler.HandleResponse(resp)

	m.hcHandler.HandleLatency(resp, latency)

	return resp, nil
}

// DefaultHTTPClientHandler describes a default http client monitoring handler
type DefaultHTTPClientHandler struct {
	requestCounter *prometheus.CounterVec
	requestLatency *prometheus.HistogramVec
}

// NewDefaultHTTPClientHandler new default http client handler
func NewDefaultHTTPClientHandler(namespace string) *DefaultHTTPClientHandler {
	handler := &DefaultHTTPClientHandler{}
	handler.Init(namespace)
	return handler
}

// Init inits default http client monitor
func (m *DefaultHTTPClientHandler) Init(namespace string) {
	//monitor := &DefaultHTTPClientHandler{}
	m.requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_client_request_counter",
			Help:      "Http client request counter",
		},
		[]string{"method", "host", "endpoint", "status"},
	)

	historyBuckets := []float64{10., 20., 30., 50., 80., 100., 200., 300., 500., 1000., 2000., 3000.}
	m.requestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_client_response_latency_millisecond",
			Help:      "Http client response latency (millisecond)",
			Buckets:   historyBuckets}, []string{"method", "host", "endpoint", "status"},
	)

	prometheus.MustRegister(m.requestCounter)
	prometheus.MustRegister(m.requestLatency)
}

// HandleRequest handles request monitoring
func (m *DefaultHTTPClientHandler) HandleRequest(r *http.Request) {
	return
}

// HandleResponse handles response monitoring
func (m *DefaultHTTPClientHandler) HandleResponse(r *http.Response) {
	if r != nil {
		m.requestCounter.
			WithLabelValues(r.Request.Method, r.Request.URL.Host, r.Request.URL.Path, fmt.Sprintf("%v", r.StatusCode)).
			Inc()
	}
}

// HandleLatency handles request latency monitoring
func (m *DefaultHTTPClientHandler) HandleLatency(r *http.Response, latency float64) {
	m.requestLatency.
		WithLabelValues(r.Request.Method, r.Request.URL.Host, r.Request.URL.Path, fmt.Sprintf("%v", r.StatusCode)).
		Observe(latency)
}
