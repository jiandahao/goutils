package monitor

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// A HTTPServerHandler is a interface used by a Monitor to handle http server monitoring handler.
type HTTPServerHandler interface {
	//	Init(namespace string, metricEndpoint string)
	HandleRequest(r *http.Request)
	HandleResponse(r *ResponseWriterInterceptor)
	HandleLatency(r *ResponseWriterInterceptor, latency float64)
}

// DefaultHTTPServerHandler describes a default http server monitor handler.
type DefaultHTTPServerHandler struct {
	requestCounter  *prometheus.CounterVec
	responseCounter *prometheus.CounterVec
	processLatency  *prometheus.HistogramVec
	metricEndpoint  string
	namespace       string
}

// NewDefaultHTTPServerHandler new default http server handler
func NewDefaultHTTPServerHandler(namespace string, metricEndpoint string) *DefaultHTTPServerHandler {
	handler := &DefaultHTTPServerHandler{}
	handler.Init(namespace, metricEndpoint)
	return handler
}

// Init inits default http server monitor
func (m *DefaultHTTPServerHandler) Init(namespace string, metricEndpoint string) {
	// monitor := &DefautlHTTPServerHandler{metricEndpoint: metricEndpoint, namespace: namespace}

	m.requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_server_request_total",
			Help:      "Http server request total",
		},
		[]string{"method", "endpoint"})

	m.responseCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_server_response_count",
			Help:      "Total http server response status counter"},
		[]string{"method", "endpoint", "status"})

	historyBuckets := []float64{10., 20., 30., 50., 80., 100., 200., 300., 500., 1000., 2000., 3000.}
	m.processLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_server_response_latency_millisecond",
			Help:      "Http server response latency (millisecond)",
			Buckets:   historyBuckets}, []string{"method", "endpoint", "status"})

	prometheus.MustRegister(m.requestCounter)
	prometheus.MustRegister(m.responseCounter)
	prometheus.MustRegister(m.processLatency)
}

// HandleRequest handles request monitoring
func (m *DefaultHTTPServerHandler) HandleRequest(r *http.Request) {
	// request counter
	m.requestCounter.WithLabelValues(r.Method, r.URL.Path).Inc()
}

// HandleResponse handles response monitoring
func (m *DefaultHTTPServerHandler) HandleResponse(wi *ResponseWriterInterceptor) {
	// response counter
	m.responseCounter.WithLabelValues(wi.Method, wi.URL.Path, fmt.Sprintf("%v", wi.statusCode))
	// TODO: parsing body and count status code in body
	_ = wi.Body
}

// HandleLatency handles request latency monitoring
func (m *DefaultHTTPServerHandler) HandleLatency(wi *ResponseWriterInterceptor, latency float64) {
	// 延迟统计
	m.processLatency.WithLabelValues(wi.Method, wi.URL.Path, fmt.Sprintf("%v", wi.statusCode)).Observe(latency)
}

// PromHTTPHandler returns prometheus http handler
func PromHTTPHandler() http.Handler {
	return promhttp.Handler()
}

// PromGinHandler returns prometheus gin handler
func PromGinHandler() gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
	})
}

// ResponseWriterInterceptor wraps ResponseWriter to get response data
type ResponseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
	URL        *url.URL
	Method     string
	Body       []byte
}

// WriteHeader overrides ResponseWriter.WriteHeader function to collect status code
func (w *ResponseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *ResponseWriterInterceptor) Write(body []byte) (int, error) {
	for _, v := range body {
		w.Body = append(w.Body, v)
	}
	return w.ResponseWriter.Write(body)
}
