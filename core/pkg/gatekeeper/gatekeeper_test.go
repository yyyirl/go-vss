// @Title        authcore
// @Description  main
// @Create       yiyiyi 2025/9/26 13:44

package gatekeeper

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/pkg/redis"
	"skeyevss/core/tps"
)

// go test -bench=. -benchtime=1000x -benchmem
func Benchmark_id_test(b *testing.B) {
	var (
		log = logx.LogConf{
			ServiceName: "gatekeeper test",
			Level:       "error",
		}

		instance = New(
			redis.NewGoRedisClient(
				"pro",
				"plain",
				tps.YamlRedis{
					Host:        "127.0.0.1:6379",
					Pass:        "",
					MaxIdle:     300,
					MaxActive:   600,
					IdleTimeout: 300,
				},
				log,
			),
			60,
			"fEaB5EYyHKsSbvWg",
			"127.0.0.1",
		)
		wg sync.WaitGroup
	)

	logx.DisableStat()
	logx.MustSetup(log)

	var num int64 = 0
	for i := 0; i < b.N; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			token, err := instance.ClientID()
			if err != nil {
				b.Errorf("生成id错误 err:%s", err)
			}

			if token != "" {
				atomic.AddInt64(&num, 1)
			}
		}(i)
	}

	go func() {
		for range time.Tick(1 * time.Second) {
			fmt.Printf("\n 生成数量: %+v \n", num)
		}
	}()

	wg.Wait()
}
