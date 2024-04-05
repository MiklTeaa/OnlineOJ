package parallelx

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"code-platform/log"
	"code-platform/pkg/errorx"
)

// Do 同步阻塞处理任务
func Do(logger *log.Logger, tasks ...func() error) error {
	var wg sync.WaitGroup
	wg.Add(len(tasks))
	errChan := make(chan error, len(tasks))
	doneChan := make(chan struct{}, 1)

	go func() {
		wg.Wait()
		doneChan <- struct{}{}
	}()

	for _, task := range tasks {
		go func(task func() error) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					err := fmt.Errorf("%v", r)
					logger.Error(err, "panic"+string(debug.Stack()))
					errChan <- errorx.InternalErr(err)
				}
			}()
			if err := task(); err != nil {
				errChan <- err
			}
		}(task)
	}

	select {
	case <-doneChan:
	case err := <-errChan:
		return err
	}
	return nil
}

// DoAsyncWithTimeOut 指定超时时间，异步非阻塞处理任务
func DoAsyncWithTimeOut(ctx context.Context, duration time.Duration, logger *log.Logger, tasks ...func(ctx context.Context) error) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%v", r)
				logger.Error(err, "panic"+string(debug.Stack()))
			}
		}()

		ctx, cancel := context.WithTimeout(ctx, duration)
		defer cancel()

		syncTasks := make([]func() error, len(tasks))
		for index := range tasks {
			task := tasks[index]
			syncTasks[index] = func() (err error) {
				if err := task(ctx); err != nil {
					return err
				}
				return nil
			}
		}

		// do sync tasks
		Do(logger, syncTasks...)
	}()
}
