/**
 * @Author:         yi
 * @Description:    main
 * @Version:        1.0.0
 * @Date:           2025/2/11 17:12
 */
package dt

import (
	"context"
	"time"
)

//	_ "net/http/pprof"
//
// // http://localhost:9100/debug/pprof/
//
//	go func() {
//		_ = http.ListenAndServe("0.0.0.0:9100", nil)
//	}()

//  go tool pprof -http=:9110 http://127.0.0.1:9100/debug/pprof/profile
//  go tool pprof -http=:9110 http://127.0.0.1:9100/debug/pprof/heap
//  go tool pprof -http=:9110 cpu.out
//  go tool pprof -http=:9110 pprof.XXX.samples.cpu.001.pb.gz

// -------
//  go tool trace trace.out
// trace.Start(os.Stderr)
// defer trace.Stop()

type debounceType struct {
	Call     func()
	ExecTime int64
}

type throttledType struct {
	Call   func()
	Cancel context.CancelFunc
}

func init() {
	go debounceRunner()
}

func SetTimeout(duration time.Duration, f func()) context.CancelFunc {
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		select {
		case <-ctx.Done():
			return

		case <-time.After(duration):
			f()
			return
		}
	}()

	return cancelFunc
}

func SetInterval(timeout int64, f func()) context.CancelFunc {
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		for {
			time.Sleep(time.Duration(timeout) * time.Second)
			select {
			case <-ctx.Done():
				return

			default:
				f()
				return
			}
		}
	}()

	return cancelFunc
}

// func Debounce(f func(), wait int) func() {
// 	var cf context.CancelFunc
// 	return func() {
// 		if cf != nil {
// 			cf()
// 		}
// 		cf = SetTimeout(f, wait)
// 	}
// }
//
// func Throttled(f func(), wait int) func() {
// 	var cf context.CancelFunc
// 	return func() {
// 		if cf == nil {
// 			cf = SetTimeout(f, wait)
// 		}
// 	}
// }
