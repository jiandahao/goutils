package monitor

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config describes monitor configuration
type Config struct {
	Namespace      string
	MetricEndpoint string
}

// Monitor describes a monitor based on prometheus
type Monitor struct {
	cfg       *Config
	hsHandler HTTPServerHandler
	hcHandler HTTPClientHandler
	dbHandler DBHandler

	transport http.RoundTripper

	collectors map[string]prometheus.Collector
}

// New new a monitor
func New(cfg *Config) *Monitor {
	monitor := Monitor{cfg: cfg, collectors: map[string]prometheus.Collector{}}
	monitor.hsHandler = NewDefaultHTTPServerHandler(cfg.Namespace, cfg.MetricEndpoint)
	monitor.hcHandler = NewDefaultHTTPClientHandler(cfg.Namespace)
	monitor.dbHandler = NewDefaultDBHandler(cfg.Namespace)
	return &monitor
}

// MustRegister registers the provided Collectors and
// panics if any error occurs.
func (m *Monitor) MustRegister(cs ...prometheus.Collector) {
	prometheus.MustRegister(cs...)
}

// Register registers the provided Collector
func (m *Monitor) Register(c prometheus.Collector) error {
	return prometheus.Register(c)
}

// RegisterWithName registers the provided Collector with specified name
func (m *Monitor) RegisterWithName(name string, c prometheus.Collector) error {
	m.collectors[name] = c
	return prometheus.Register(c)
}

// NamedCollector returns collector with specified name
func (m *Monitor) NamedCollector(name string) prometheus.Collector {
	c, ok := m.collectors[name]
	if ok {
		return c
	}

	return nil
}

// HTTPHandler returns an http.Handler for the prometheus Gatherer
func (m *Monitor) HTTPHandler() http.Handler {
	return promhttp.Handler()
}

// SetHTTPServerHandler sets http server monitoring handler
func (m *Monitor) SetHTTPServerHandler(handler HTTPServerHandler) {
	m.hsHandler = handler
}

// SetHTTPClientHandler sets http client monitoring handler
func (m *Monitor) SetHTTPClientHandler(handler HTTPClientHandler) {
	m.hcHandler = handler
}

// SetDBHandler sets db monitoring handler
func (m *Monitor) SetDBHandler(handler DBHandler) {
	m.dbHandler = handler
}

// WrapAndServeHTTPServer wraps http server with promethues collectors
func (m *Monitor) WrapAndServeHTTPServer(srv *http.Server) {
	handler := srv.Handler
	srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		endpoint := r.URL.Path

		if m.cfg.MetricEndpoint != "" {
			if method == http.MethodGet && endpoint == m.cfg.MetricEndpoint {
				handler := promhttp.Handler()
				handler.ServeHTTP(w, r)
				return
			}
		}

		// handle http request
		m.hsHandler.HandleRequest(r)

		wi := &ResponseWriterInterceptor{
			ResponseWriter: w,
			URL:            r.URL,
			Method:         r.Method,
			statusCode:     http.StatusOK,
		}
		start := time.Now()
		handler.ServeHTTP(wi, r)
		latency := (float64)(time.Since(start) / time.Millisecond)

		// count latency
		m.hsHandler.HandleLatency(wi, latency)

		// count response status
		m.hsHandler.HandleResponse(wi)
	})
}

// WrapAndServeHTTPClient wraps http client transport with monitoring transport.
func (m *Monitor) WrapAndServeHTTPClient(client *http.Client) {

	assertNotNil(client, "invalid http client; could not be nil")

	clientTrans := client.Transport
	if clientTrans == nil {
		clientTrans = http.DefaultTransport
	}

	client.Transport = &RoundTripWrapper{
		hcHandler: m.hcHandler,
		transport: clientTrans,
	}
}

// WrapAndServeDB warps db connection to get execution statistics
func (m *Monitor) WrapAndServeDB(conn *gorm.DB) {
	assertNotNil(conn, "invalid gorm.DB instance")

	conn.Callback().Create().Before("gorm:create").Register("monitor:before_create", m.dbHandler.BeforeExecution)
	conn.Callback().Create().After("gorm:create").Register("monitor:after_create", m.dbHandler.AfterExecution)

	conn.Callback().Delete().Before("gorm:delete").Register("monitor:before_delete", m.dbHandler.BeforeExecution)
	conn.Callback().Delete().After("gorm:delete").Register("monitor:after_delete", m.dbHandler.AfterExecution)

	conn.Callback().Update().Before("gorm:update").Register("monitor:before_update", m.dbHandler.BeforeExecution)
	conn.Callback().Update().After("gorm:update").Register("monitor:after_update", m.dbHandler.AfterExecution)

	conn.Callback().Query().Before("gorm:query").Register("monitor:before_query", m.dbHandler.BeforeExecution)
	conn.Callback().Query().After("gorm:query").Register("monitor:after_query", m.dbHandler.AfterExecution)

	conn.Callback().RowQuery().Before("gorm:row_query").Register("monitor:before_rowquery", m.dbHandler.BeforeExecution)
	conn.Callback().RowQuery().After("gorm:row_query").Register("monitor:after_rowquery", m.dbHandler.AfterExecution)
}

// ListenAndServe start a http server that expose metrics endpoint
func (m *Monitor) ListenAndServe(addr string) error {
	mux := http.DefaultServeMux
	mux.Handle(m.cfg.MetricEndpoint, promhttp.Handler())

	return http.ListenAndServe(addr, mux)
}

// Handler returns an http.Handler for monitoring.
func (m *Monitor) Handler() http.Handler {
	return promhttp.Handler()
}
