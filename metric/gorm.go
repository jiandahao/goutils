package metric

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	promPlugin "gorm.io/plugin/prometheus"
)

// NewGORMMetricPlugin creates a prom plugin that capable to collect mysql system metrics,
// including threads_running, threads_created, threads_connected and slow_quries.
func NewGORMMetricPlugin(schema string) *promPlugin.Prometheus {
	p := promPlugin.New(promPlugin.Config{
		DBName:          schema, // `DBName` as metrics label
		RefreshInterval: 15,     // refresh metrics interval (default 15 seconds)
		MetricsCollector: []promPlugin.MetricsCollector{
			&promPlugin.MySQL{
				Prefix:        fmt.Sprintf("%s_gorm_status_", metricNamespace),
				VariableNames: []string{"Threads_running", "Threads_created", "Threads_connected", "Slow_queries"},
			},
		},
	})

	p.Labels = commonLables

	return p
}

// GORMLoggerWithMetircCollector is a implementation of gorm/logger.Interface which is capable to
// print sql messages and collect execution metrics.
type GORMLoggerWithMetircCollector struct {
	logger.Interface

	counter *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

// NewGORMLoggerWithMetircCollector returns a GORMLoggerWithMetircCollector instances.
func NewGORMLoggerWithMetircCollector(schema string, l logger.Interface) logger.Interface {
	c := &GORMLoggerWithMetircCollector{
		Interface: l,
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: metricNamespace,
				Subsystem: "db_request",
				Name:      "total",
				Help:      "db request total",
			},
			[]string{"caller", "status"},
		),
		latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: metricNamespace,
				Subsystem: "db_response",
				Name:      "milliseconds",
				Help:      "db response latency (milliseconds)",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"caller"},
		),
	}

	defaultRegister.MustRegister(c.counter, c.latency)
	return c
}

// Trace print sql message and collect metrics.
func (w *GORMLoggerWithMetircCollector) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin) / time.Millisecond
	caller := path.Base(utils.FileWithLineNum())

	w.latency.WithLabelValues(caller).Observe(float64(elapsed))

	switch {
	case err != nil && (!errors.Is(err, gorm.ErrRecordNotFound)):
		w.counter.WithLabelValues(caller, err.Error()).Inc()
	default:
		w.counter.WithLabelValues(caller, "ok").Inc()
	}

	w.Interface.Trace(ctx, begin, fc, err)
}
