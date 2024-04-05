package monitor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	LabelAPI        = "api"
	LabelMethod     = "method"
	LabelStatusCode = "status_code"
)

var (
	ServerhandleCounterCollector = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "server_http_throughput",
			Help: "Total number of HTTP transaction completed by the server",
		},
		[]string{LabelAPI, LabelMethod, LabelStatusCode},
	)

	ServerHandleLatencyCollector = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "server_latency_us",
			Help:    "Latency (microseconds) of HTTP that had been handled by the server",
			Buckets: []float64{10, 20, 40, 80, 160, 320, 640, 1280, 2560, 5120, 10240, 20480, 51200},
		},
		[]string{LabelAPI, LabelMethod, LabelStatusCode},
	)
)

func init() {
	prometheus.MustRegister(ServerhandleCounterCollector, ServerHandleLatencyCollector)
}

func Init() {
	engine := gin.New()
	engine.Use(gin.Recovery())
	metricsHandler := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
		ErrorHandling:     promhttp.ContinueOnError,
	})
	engine.GET("/metrics", func(c *gin.Context) {
		metricsHandler.ServeHTTP(c.Writer, c.Request)
	})
	go func() {
		if err := http.ListenAndServe(":8090", engine); err != nil {
			panic(err)
		}
	}()
}
