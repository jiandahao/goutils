package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

var defaultRegister prometheus.Registerer = prometheus.DefaultRegisterer
var metricNamespace string = "default"
var commonLables prometheus.Labels = make(prometheus.Labels)

// SetDefaultRegister overrides the default register.
func SetDefaultRegister(reg prometheus.Registerer) {
	defaultRegister = reg
}

// MustRegister register collectors.
func MustRegister(cs ...prometheus.Collector) {
	defaultRegister.MustRegister(cs...)
}

// SetNamespace set metric namespace
func SetNamespace(namespace string) {
	metricNamespace = namespace
}

// SetCommonLabels set common labels
func SetCommonLabels(labels prometheus.Labels) {
	commonLables = labels
}
