package metric

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type httpTransport struct {
	clientType   string
	transport    http.RoundTripper
	collector    *httpClientCollector
	errorHandler errorExtractor
}

var httpClientCollectorMux sync.Mutex
var httpClientCollectors = map[string]*httpClientCollector{}

type httpClientCollector struct {
	requestCounter *prometheus.CounterVec
	requestLatency *prometheus.HistogramVec
}

func getOrRegisterHTTPClientCollector() *httpClientCollector {
	httpClientCollectorMux.Lock()
	defer httpClientCollectorMux.Unlock()

	if c, ok := httpClientCollectors[metricNamespace]; ok && c != nil {
		return c
	}

	var lvls = []string{
		"method",      // http method，GET / POST etc.
		"host",        // host
		"endpoint",    // request path
		"status_code", // http status code, 200 etc.
		"status",      // request status represents the error details if request failed.
		"client_type", // client type
	}

	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Subsystem: "http_client_request",
			Name:      "total",
			Help:      "Http client request total",
		},
		lvls,
	)

	requestLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricNamespace,
			Subsystem: "http_client_request",
			Name:      "second",
			Help:      "Http client request latency (second)",
			Buckets:   prometheus.DefBuckets,
		},
		lvls,
	)

	defaultRegister.MustRegister(requestCounter)
	defaultRegister.MustRegister(requestLatency)

	c := &httpClientCollector{
		requestCounter: requestCounter,
		requestLatency: requestLatency,
	}

	httpClientCollectors[metricNamespace] = c
	return c
}

// errorExtractor represents a handler for extracting biz error message of specified client.
type errorExtractor func(responseBody []byte) string

// NewHTTPTransport creates a transport injected with a metrics collector.
func NewHTTPTransport(clientName string, transport http.RoundTripper, ee errorExtractor) http.RoundTripper {
	collector := getOrRegisterHTTPClientCollector()
	if transport == nil {
		transport = http.DefaultTransport
	}

	ht := &httpTransport{
		collector:    collector,
		transport:    transport,
		clientType:   clientName,
		errorHandler: ee,
	}

	return ht
}

// RoundTrip executes a single HTTP transaction, returning
// a Response for the provided Request. It is used by HTTP client monitor.
func (m *httpTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := m.transport.RoundTrip(r)
	if err != nil {
		m.collector.requestCounter.WithLabelValues(r.Method, r.Host, r.URL.Path, "", err.Error(), m.clientType).Inc()
		return nil, err
	}

	elapsed := float64(time.Since(start)) / float64(time.Second)
	var status string
	if m.errorHandler != nil {
		var val []byte
		resp.Body, val = shadowRead(resp.Body)
		status = m.errorHandler(val)
	}

	if len(status) > 60 {
		status = status[:60]
	}

	labelValues := []string{r.Method, r.Host, r.URL.Path, fmt.Sprint(resp.StatusCode), status, m.clientType}

	m.collector.requestCounter.WithLabelValues(labelValues...).Inc()
	m.collector.requestLatency.WithLabelValues(labelValues...).Observe(elapsed)

	return resp, nil
}

// NewHTTPClient wraps a http.Client with a metric collector.
func NewHTTPClient(clientName string, c *http.Client) *http.Client {
	client := c
	if client == nil {
		client = &http.Client{Transport: http.DefaultTransport}
	}

	clientTrans := client.Transport

	client.Transport = NewHTTPTransport(clientName, clientTrans, nil)

	return client
}

func shadowRead(reader io.ReadCloser) (io.ReadCloser, []byte) {
	val, err := ioutil.ReadAll(reader)
	if err != nil {
		return reader, nil
	}
	return ioutil.NopCloser(bytes.NewBuffer(val)), val
}
