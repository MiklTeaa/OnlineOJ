package monitor

import "github.com/prometheus/client_golang/prometheus"

var (
	IDESweaterCollector = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ide_sweater_duration_milliseconds",
		Help: "ide_sweater work latency distributions.",
	})
)

func init() {
	prometheus.MustRegister(IDESweaterCollector)
}
