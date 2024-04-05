package monitor

import (
	"code-platform/log"
	"fmt"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	goroutineCollector = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "goroutine_nums",
		Help: "Current num of goroutines",
	})
)

func init() {
	initForGoroutineCollector()
	prometheus.MustRegister(goroutineCollector)
}

func initForGoroutineCollector() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%v", r)
				log.Errorf(err, "panic")
				time.Sleep(time.Minute)
				initForGoroutineCollector()
			}
		}()
		// 15秒统计一次 Goroutine 数量
		for ; ; time.Sleep(time.Second * 15) {
			goroutineCollector.Set(float64(runtime.NumGoroutine()))
		}
	}()
}
