package monitor

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/prometheus/client_golang/prometheus"
)

// A DBHandler is a interface used by a Monitor to handle database monitoring handler
type DBHandler interface {
	//Init(namespace string)
	BeforeExecution(scope *gorm.Scope)
	AfterExecution(scope *gorm.Scope)
}

// DefaultDBHandler describes a default db monitoring handler
type DefaultDBHandler struct {
	dbProcessLatency *prometheus.HistogramVec
	dbCounter        *prometheus.CounterVec
	dbErrorCounter   *prometheus.CounterVec

	namespace string
}

// NewDefaultDBHandler new default db handler
func NewDefaultDBHandler(namespace string) *DefaultDBHandler {
	handler := &DefaultDBHandler{}
	handler.Init(namespace)
	return handler
}

// Init inits default db handler
func (m *DefaultDBHandler) Init(namespace string) {
	historyBuckets := []float64{10., 20., 30., 50., 80., 100., 200., 300., 500., 1000., 2000., 3000.}
	m.dbProcessLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "db_latency_millisecond",
			Help:      "DB latency (millisecond)",
			Buckets:   historyBuckets,
		},
		[]string{"tablename", "sql"},
	)

	m.dbCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "db_counter",
			Help:      "DB counter",
		},
		[]string{"tablename", "sql"},
	)

	m.dbErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "db_error_count",
			Help:      "DB error count",
		},
		[]string{"tablename", "sql"},
	)

	prometheus.MustRegister(m.dbCounter)
	prometheus.MustRegister(m.dbErrorCounter)
	prometheus.MustRegister(m.dbProcessLatency)
}

// BeforeExecution is before callback, prepares some useful data for monitoring system
func (m *DefaultDBHandler) BeforeExecution(scope *gorm.Scope) {
	scope.Set("start_at", time.Now())
}

// AfterExecution calculates the execution details
func (m *DefaultDBHandler) AfterExecution(scope *gorm.Scope) {
	startAt, ok := scope.Get("start_at")
	if !ok {
		return
	}

	start, ok := startAt.(time.Time)
	if !ok {
		fmt.Println("Type assertion failure in After callback; falied to get start time")
		return
	}
	latency := (float64)(time.Since(start) / time.Millisecond)

	tableName := scope.TableName()

	m.dbCounter.WithLabelValues(tableName, scope.SQL).Inc()

	m.dbProcessLatency.WithLabelValues(tableName, scope.SQL).Observe(latency)

	if scope.HasError() {
		m.dbErrorCounter.WithLabelValues(tableName, scope.SQL).Inc()
		return
	}
}
