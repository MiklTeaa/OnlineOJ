package main

import (
	"code-platform/api/http/handler"
	"code-platform/monitor"
	"code-platform/pkg/notifyx"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// receiving signal to close storage
	notifyx.Notify()

	// prometheus monitor
	monitor.Init()

	handler.Init()
}
