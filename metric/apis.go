package metric

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PromGinHandler returns prometheus gin handler
func PromGinHandler() gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
		promhttp.InstrumentMetricHandler(
			defaultRegister, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}),
		).ServeHTTP(ctx.Writer, ctx.Request)
	})
}

// GinAPICollector returns a gin.HandlerFunc that collect metrics.
func GinAPICollector(metricPath string, codeExtractor func(respBody []byte) string) gin.HandlerFunc {
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Subsystem: "http_requests",
			Name:      "total",
			Help:      "Http server request total",
		},
		[]string{"method", "endpoint", "status", "code"},
	)

	processLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricNamespace,
			Subsystem: "http_requests",
			Name:      "second",
			Help:      "Http server response latency (second)",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	defaultRegister.MustRegister(requestCounter)
	defaultRegister.MustRegister(processLatency)

	return func(c *gin.Context) {
		method := c.Request.Method
		endpoint := c.Request.URL.Path

		if endpoint == metricPath {
			c.Next()
			return
		}

		wi := &ResponseWriterInterceptor{
			ResponseWriter: c.Writer,
			statusCode:     http.StatusOK,
		}

		c.Writer = wi

		startAt := time.Now()
		c.Next()
		elapsed := float64(time.Since(startAt)) / float64(time.Second)

		// observe latency
		processLatency.WithLabelValues(method, endpoint, fmt.Sprintf("%v", wi.statusCode)).Observe(elapsed)

		var code string
		if codeExtractor != nil {
			code = codeExtractor(wi.Body)
		}

		requestCounter.WithLabelValues(method, endpoint, fmt.Sprint(wi.statusCode), fmt.Sprint(code)).Inc()
	}
}

// DefaultAPICodeExtractor for code-message formated response
var DefaultAPICodeExtractor = func(respBody []byte) string {
	// parsing body and extract status code
	var status commonResponse
	_ = json.Unmarshal(respBody, &status)

	return fmt.Sprint(status.Code)
}

type commonResponse struct {
	Code int `json:"code"`
}

// ResponseWriterInterceptor wraps gin.ResponseWriter to get response data.
type ResponseWriterInterceptor struct {
	gin.ResponseWriter
	statusCode int
	Body       []byte
}

// WriteHeader overrides ResponseWriter.WriteHeader function to collect status code.
func (w *ResponseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write overrides ResponseWriter.Write to collect original response body.
func (w *ResponseWriterInterceptor) Write(body []byte) (int, error) {
	w.Body = append(w.Body, body...)
	return w.ResponseWriter.Write(body)
}
