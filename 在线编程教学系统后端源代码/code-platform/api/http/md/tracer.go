package md

import (
	"strconv"
	"time"

	"code-platform/log"
	"code-platform/monitor"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

func Tracer(apiName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		labels := prometheus.Labels{
			monitor.LabelAPI:        apiName,
			monitor.LabelMethod:     c.Request.Method,
			monitor.LabelStatusCode: strconv.Itoa(c.Writer.Status()),
		}
		// monitor for counter
		counter, err := monitor.ServerhandleCounterCollector.GetMetricWith(labels)
		if err != nil {
			log.Errorf(err, "monitor for http handler %q with method %q counter failed", c.Request.URL.Path, c.Request.Method)
			return
		}
		counter.Inc()

		// monitor for latency
		observer, err := monitor.ServerHandleLatencyCollector.GetMetricWith(labels)
		if err != nil {
			log.Errorf(err, "monitor for http handler %q with method %q latency observer failed", c.Request.URL.Path, c.Request.Method)
			return
		}

		observer.Observe(float64(time.Since(startTime) / time.Microsecond))
	}
}
