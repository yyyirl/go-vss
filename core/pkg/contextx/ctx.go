/**
 * @Author:         yi
 * @Description:    ctx
 * @Version:        1.0.0
 * @Date:           2022/12/7 11:18
 */
package contextx

import (
	"context"
	"time"
)

type Ctx struct {
	Context   context.Context
	CtxCancel func()
}

func NewCtx(ctx context.Context) *Ctx {
	ctx, cancel := context.WithCancel(ctx)

	return &Ctx{
		Context: ctx,
		CtxCancel: func() {
			cancel()
		},
	}
}

func NewCtxWithDuration(t time.Duration) *Ctx {
	var ctx = context.Background()
	ctx, cancel := context.WithTimeout(ctx, t)

	return &Ctx{
		Context: ctx,
		CtxCancel: func() {
			cancel()
		},
	}
}

func NewCtxWith(ctx context.Context, t time.Duration) *Ctx {
	ctx, cancel := context.WithTimeout(ctx, t)

	return &Ctx{
		Context: ctx,
		CtxCancel: func() {
			cancel()
		},
	}
}

func CtxWatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// LogInfo("exec context done")
			return

		default:
			time.Sleep(1 * time.Second)
			// fmt.Println("goroutine time=", time.Now().Unix())
		}
	}
}
