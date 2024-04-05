package notifyx

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"code-platform/log"
	"code-platform/repository"
)

func Notify() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%v", r)
				log.Errorf(err, "panic in notify")
			}
		}()

		log.Sub("notify").Info("begin to start listen signal")
		sig := <-c
		switch sig {
		case syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT:
			if repository.Storage != nil {
				log.Sub("notify").Infof("storage close for receiving signal %v", sig)
				repository.Storage.Close()
				os.Exit(0)
			}
		}
	}()
}
