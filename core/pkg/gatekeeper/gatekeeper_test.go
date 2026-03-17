// @Title        gatekeeper
// @Description  测试用例
// @Create       yiyiyi 2025/9/26 13:44

package gatekeeper

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/redis"
	"skeyevss/core/tps"
)

// go test -bench=. -benchtime=1000x -benchmem

// 测试初始化
func setupTest(_ *testing.T) *Gatekeeper {
	var log = logx.LogConf{
		ServiceName: "gatekeeper test",
		Level:       "error",
	}

	logx.DisableStat()
	logx.MustSetup(log)

	return New(
		redis.NewGoRedisClient(
			"pro",
			"plain",
			tps.YamlRedis{
				Host: "127.0.0.1:6379",
				// Pass:        "",
				MaxIdle:     300,
				MaxActive:   600,
				IdleTimeout: 300,
			},
			log,
		),
		60000,
		"fEaB5EYyHKsSbvWg",
		"127.0.0.1",
	)
}

// 清理测试数据
func cleanupTest(t *testing.T, g *Gatekeeper) {
	// 清理Redis中的数据
	var keys = []string{
		idCacheKey,
		uniqueIdsCacheKey,
	}
	for _, key := range keys {
		_, err := g.RedisClient.Del(key)
		if err != nil {
			t.Logf("cleanup key %s error: %v", key, err)
		}
	}

	// 清理黑名单相关的key
	var (
		pattern          = blacklistKey + ":*"
		blacklistKeys, _ = g.RedisClient.Keys(pattern)
	)
	if len(blacklistKeys) > 0 {
		_, _ = g.RedisClient.Del(blacklistKeys...)
	}

	// 清理限流相关的key
	var (
		ratePattern = rateLimitKey + ":*"
		rateKeys, _ = g.RedisClient.Keys(ratePattern)
	)
	if len(rateKeys) > 0 {
		_, _ = g.RedisClient.Del(rateKeys...)
	}

	g.Stop()
}

// ==================== 单元测试 ====================

// TestClientID 测试生成客户端ID
func TestClientID(t *testing.T) {
	var g = setupTest(t)
	defer cleanupTest(t, g)

	t.Run("生成单个ID", func(t *testing.T) {
		id, err := g.ClientID()
		if err != nil {
			t.Fatalf("生成ID失败: %v", err)
		}

		if id == "" {
			t.Fatal("生成的ID为空")
		}

		t.Logf("生成的ID: %s", id)
	})

	t.Run("生成多个ID检查唯一性", func(t *testing.T) {
		var (
			idMap = make(map[string]bool)
			count = 1000
		)
		for i := 0; i < count; i++ {
			id, err := g.ClientID()
			if err != nil {
				t.Fatalf("生成ID失败: %v", err)
			}

			if idMap[id] {
				t.Fatalf("ID重复: %s", id)
			}
			idMap[id] = true
		}

		if len(idMap) != count {
			t.Fatalf("ID数量不匹配: 期望 %d, 实际 %d", count, len(idMap))
		}
		t.Logf("成功生成 %d 个唯一ID", count)
	})
}

// TestCredential 测试生成凭证
func TestCredential(t *testing.T) {
	var g = setupTest(t)
	defer cleanupTest(t, g)

	t.Run("为ID生成凭证", func(t *testing.T) {
		// 先获取ID
		id, err := g.ClientID()
		if err != nil {
			t.Fatalf("生成ID失败: %v", err)
		}

		// 生成凭证
		token, err := g.Credential(id)
		if err != nil {
			t.Fatalf("生成凭证失败: %v", err)
		}
		if token == "" {
			t.Fatal("生成的凭证为空")
		}
		t.Logf("ID: %s 的凭证: %s", id, token)
	})
}

// TestGuard 测试访问守卫
func TestGuard(t *testing.T) {
	var g = setupTest(t)
	defer cleanupTest(t, g)

	t.Run("有效的凭证访问", func(t *testing.T) {
		// 生成ID和凭证
		id, _ := g.ClientID()
		token, _ := g.Credential(id)

		// 访问守卫
		newToken, err := g.Guard(token)
		if err != nil {
			t.Fatalf("守卫验证失败: %v", err)
		}

		if newToken != "" {
			t.Logf("凭证已续期，新token: %s", newToken)
		}
	})

	t.Run("无效的凭证", func(t *testing.T) {
		_, err := g.Guard("invalid_token")
		if err == nil {
			t.Fatal("无效凭证应该返回错误")
		}
		t.Logf("预期的错误: %v", err)
	})

	t.Run("空凭证", func(t *testing.T) {
		_, err := g.Guard("")
		if err == nil {
			t.Fatal("空凭证应该返回错误")
		}
		t.Logf("预期的错误: %v", err)
	})

	t.Run("不存在的凭证", func(t *testing.T) {
		// 生成一个有效的token但不存储
		var (
			id, _ = g.ClientID()
			item  = &CacheItem{
				ID:     id,
				Expire: uint64(time.Now().UnixMilli()) + g.Expire,
			}
			b, _       = functions.JSONMarshal(item)
			encrypt, _ = functions.NewCrypto([]byte(g.Key)).Encrypt(b)
			fakeToken  = prefix + g.swapStringParts(encrypt) + suffix
		)
		_, err := g.Guard(fakeToken)
		if err == nil {
			t.Fatal("不存在的凭证应该返回错误")
		}

		if err.Error() != "credential not found" {
			t.Fatalf("错误信息不匹配: 期望 'credential not found', 实际 '%v'", err)
		}
	})
}

// TestBlacklist 测试黑名单功能
func TestBlacklist(t *testing.T) {
	var g = setupTest(t)
	defer cleanupTest(t, g)

	t.Run("加入黑名单", func(t *testing.T) {
		var (
			id       = "test_blacklist_id"
			reason   = "测试黑名单"
			duration = 10 * time.Second
		)
		// 加入黑名单
		if err := g.AddToBlacklist(id, reason, duration); err != nil {
			t.Fatalf("加入黑名单失败: %v", err)
		}

		// 检查是否在黑名单中
		blacklisted, item, err2 := g.IsBlacklisted(id)
		if err2 != nil {
			t.Fatalf("检查黑名单失败: %v", err2)
		}

		if !blacklisted {
			t.Fatal("ID应该在黑名单中")
		}

		if item.Reason != reason {
			t.Fatalf("原因不匹配: 期望 %s, 实际 %s", reason, item.Reason)
		}
	})

	t.Run("从黑名单移除", func(t *testing.T) {
		var id = "test_remove_id"
		// 加入黑名单
		_ = g.AddToBlacklist(id, "test", 1*time.Hour)
		// 从黑名单移除
		if err := g.RemoveFromBlacklist(id); err != nil {
			t.Fatalf("移除黑名单失败: %v", err)
		}

		// 检查是否已移除
		if blacklisted, _, _ := g.IsBlacklisted(id); blacklisted {
			t.Fatal("ID不应该在黑名单中")
		}
	})

	t.Run("黑名单过期", func(t *testing.T) {
		var (
			id       = "test_expire_id"
			duration = 1 * time.Second
		)
		// 加入黑名单
		_ = g.AddToBlacklist(id, "expire test", duration)
		// 等待过期
		time.Sleep(duration + 500*time.Millisecond)
		// 检查是否自动过期
		blacklisted, _, err := g.IsBlacklisted(id)
		if err != nil {
			t.Fatalf("检查黑名单失败: %v", err)
		}

		if blacklisted {
			t.Fatal("黑名单应该已过期")
		}
	})

	t.Run("黑名单与凭证访问", func(t *testing.T) {
		// 生成ID和凭证
		var (
			id, _    = g.ClientID()
			token, _ = g.Credential(id)
		)
		// 加入黑名单
		_ = g.AddToBlacklist(id, "blocked", 1*time.Hour)

		// 尝试访问
		_, err := g.Guard(token)
		if err == nil {
			t.Fatal("黑名单中的ID应该被拒绝访问")
		}

		if !strings.Contains(err.Error(), "blacklisted") {
			t.Fatalf("错误信息应该包含blacklisted, 实际: %v", err)
		}

		t.Logf("预期的拒绝访问: %v", err)
	})
}

// TestRateLimit 测试限流功能
func TestRateLimit(t *testing.T) {
	var g = setupTest(t)
	defer cleanupTest(t, g)

	g.SetRateLimit(5)

	t.Run("基础限流测试", func(t *testing.T) {
		var (
			id           = "rate_limit_test_id"
			allowedCount = 0
			deniedCount  = 0
		)
		// 尝试10次请求
		for i := 0; i < 10; i++ {
			allowed, current, err := g.RateLimit(id)
			if err != nil {
				t.Fatalf("限流检查失败: %v", err)
			}

			if allowed {
				allowedCount++
			} else {
				deniedCount++
			}

			t.Logf("请求 %d: allowed=%v, current=%d", i+1, allowed, current)
		}

		if allowedCount != 5 {
			t.Fatalf("允许次数不匹配: 期望 5, 实际 %d", allowedCount)
		}

		if deniedCount != 5 {
			t.Fatalf("拒绝次数不匹配: 期望 5, 实际 %d", deniedCount)
		}
	})

	t.Run("限流状态查询", func(t *testing.T) {
		var id = "rate_limit_status_id"
		for i := 0; i < 3; i++ {
			_, _, _ = g.RateLimit(id)
		}

		current, limit, resetTime, err := g.GetRateLimitStatus(id)
		if err != nil {
			t.Fatalf("获取限流状态失败: %v", err)
		}

		if current != 3 {
			t.Fatalf("当前请求数不匹配: 期望 3, 实际 %d", current)
		}

		if limit != 5 {
			t.Fatalf("限制数不匹配: 期望 5, 实际 %d", limit)
		}

		if resetTime.Before(time.Now()) {
			t.Fatal("重置时间应该在当前时间之后")
		}

		t.Logf("限流状态: current=%d, limit=%d, reset=%v", current, limit, resetTime)
	})

	t.Run("不同ID独立限流", func(t *testing.T) {
		var (
			id1 = "id1"
			id2 = "id2"
		)
		for i := 0; i < 6; i++ {
			g.RateLimit(id1)
		}

		for i := 0; i < 3; i++ {
			g.RateLimit(id2)
		}

		current1, _, _, _ := g.GetRateLimitStatus(id1)
		if current1 != 5 { // 应该被限流，最多5次
			t.Fatalf("id1请求数不匹配: 期望 5, 实际 %d", current1)
		}

		current2, _, _, _ := g.GetRateLimitStatus(id2)
		if current2 != 3 {
			t.Fatalf("id2请求数不匹配: 期望 3, 实际 %d", current2)
		}
	})

	t.Run("限流自动重置", func(t *testing.T) {
		var id = "rate_limit_reset_id"
		for i := 0; i < 5; i++ {
			g.RateLimit(id)
		}

		// 第6次应该被拒绝
		allowed, _, _ := g.RateLimit(id)
		if allowed {
			t.Fatal("第6次请求应该被拒绝")
		}

		// 等待1秒窗口重置
		time.Sleep(1*time.Second + 100*time.Millisecond)

		// 再次请求应该允许
		allowed, _, _ = g.RateLimit(id)
		if !allowed {
			t.Fatal("窗口重置后应该允许请求")
		}
	})

	t.Run("Guard集成限流", func(t *testing.T) {
		g.SetRateLimit(2)

		// 生成ID和凭证
		var (
			id, _    = g.ClientID()
			token, _ = g.Credential(id)
		)
		// 第一次访问
		_, err1 := g.Guard(token)
		if err1 != nil {
			t.Fatalf("第一次访问失败: %v", err1)
		}

		// 第二次访问
		_, err2 := g.Guard(token)
		if err2 != nil {
			t.Fatalf("第二次访问失败: %v", err2)
		}

		// 第三次访问应该被限流
		_, err3 := g.Guard(token)
		if err3 == nil {
			t.Fatal("第三次访问应该被限流")
		}

		if !strings.Contains(err3.Error(), "rate limit exceeded") {
			t.Fatalf("错误信息应该包含rate limit exceeded, 实际: %v", err3)
		}
		t.Logf("预期的限流错误: %v", err3)
	})
}

// TestTokenRenewal 测试凭证续期
func TestTokenRenewal(t *testing.T) {
	var g = setupTest(t)
	defer cleanupTest(t, g)

	// 设置较短的有效期便于测试
	g.Expire = 2000 // 2秒
	t.Run("半生命周期续期", func(t *testing.T) {
		// 生成ID和凭证
		var (
			id, _    = g.ClientID()
			token, _ = g.Credential(id)
		)
		// 等待超过一半有效期
		time.Sleep(1100 * time.Millisecond) // 1.1秒

		// 访问守卫，应该触发续期
		newToken, err := g.Guard(token)
		if err != nil {
			t.Fatalf("守卫验证失败: %v", err)
		}

		if newToken == "" {
			t.Fatal("应该返回续期后的新token")
		}

		if newToken == token {
			t.Fatal("新token不应该与旧token相同")
		}

		t.Logf("续期成功: 旧token=%s..., 新token=%s...", token[:20], newToken[:20])
	})

	t.Run("有效期初期不续期", func(t *testing.T) {
		// 生成ID和凭证
		var (
			id, _    = g.ClientID()
			token, _ = g.Credential(id)
		)
		// 立即访问，不应该续期
		newToken, err := g.Guard(token)
		if err != nil {
			t.Fatalf("守卫验证失败: %v", err)
		}
		if newToken != "" {
			t.Fatal("有效期初期不应该续期")
		}
	})
}

// TestConcurrency 测试并发场景
func TestConcurrency(t *testing.T) {
	var g = setupTest(t)
	defer cleanupTest(t, g)

	t.Run("并发生成ID", func(t *testing.T) {
		var (
			wg         sync.WaitGroup
			count      = 100
			idMap      = sync.Map{}
			errorCount int32
		)
		for i := 0; i < count; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				id, err := g.ClientID()
				if err != nil {
					atomic.AddInt32(&errorCount, 1)
					t.Errorf("生成ID失败: %v", err)
					return
				}

				if _, loaded := idMap.LoadOrStore(id, true); loaded {
					atomic.AddInt32(&errorCount, 1)
					t.Errorf("ID重复: %s", id)
				}
			}()
		}

		wg.Wait()

		if errorCount > 0 {
			t.Fatalf("并发测试有 %d 个错误", errorCount)
		}

		// 验证生成的ID数量
		var actualCount = 0
		idMap.Range(func(_, _ interface{}) bool {
			actualCount++
			return true
		})

		if actualCount != count {
			t.Fatalf("ID数量不匹配: 期望 %d, 实际 %d", count, actualCount)
		}
		t.Logf("成功并发生成 %d 个唯一ID", actualCount)
	})

	t.Run("并发访问守卫", func(t *testing.T) {
		// 生成一个ID和凭证
		var (
			id, _    = g.ClientID()
			token, _ = g.Credential(id)

			wg           sync.WaitGroup
			count        = 50
			successCount int32
			errorCount   int32
		)
		for i := 0; i < count; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				_, err := g.Guard(token)
				if err != nil {
					atomic.AddInt32(&errorCount, 1)
				} else {
					atomic.AddInt32(&successCount, 1)
				}
			}()
		}

		wg.Wait()

		t.Logf("并发访问结果: 成功=%d, 失败=%d", successCount, errorCount)

		// 由于限流，应该有一部分成功一部分失败
		if successCount == 0 {
			t.Fatal("没有成功的请求")
		}
	})
}

// ==================== 基准测试 ====================

// BenchmarkClientID 基准测试：生成客户端ID
func BenchmarkClientID(b *testing.B) {
	var g = setupTest(nil)
	defer cleanupTest(nil, g)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := g.ClientID()
		if err != nil {
			b.Fatalf("生成ID失败: %v", err)
		}
	}
}

// BenchmarkCredential 基准测试：生成凭证
func BenchmarkCredential(b *testing.B) {
	var g = setupTest(nil)
	defer cleanupTest(nil, g)

	// 预先生成ID
	var id, _ = g.ClientID()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := g.Credential(id)
		if err != nil {
			b.Fatalf("生成凭证失败: %v", err)
		}
	}
}

// BenchmarkGuard 基准测试：访问守卫
func BenchmarkGuard(b *testing.B) {
	var g = setupTest(nil)
	defer cleanupTest(nil, g)

	// 预先生成ID和凭证
	var id, _ = g.ClientID()
	var token, _ = g.Credential(id)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := g.Guard(token)
		if err != nil && err.Error() != "rate limit exceeded" {
			b.Fatalf("守卫验证失败: %v", err)
		}
	}
}

// BenchmarkRateLimit 基准测试：限流检查
func BenchmarkRateLimit(b *testing.B) {
	var g = setupTest(nil)
	defer cleanupTest(nil, g)

	var id = "bench_rate_limit_id"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := g.RateLimit(id)
		if err != nil {
			b.Fatalf("限流检查失败: %v", err)
		}
	}
}

// BenchmarkBlacklist 基准测试：黑名单检查
func BenchmarkBlacklist(b *testing.B) {
	var g = setupTest(nil)
	defer cleanupTest(nil, g)

	var id = "bench_blacklist_id"
	_ = g.AddToBlacklist(id, "benchmark", 1*time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := g.IsBlacklisted(id)
		if err != nil {
			b.Fatalf("黑名单检查失败: %v", err)
		}
	}
}

// BenchmarkConcurrentID 并发基准测试：生成ID
func BenchmarkConcurrentID(b *testing.B) {
	var g = setupTest(nil)
	defer cleanupTest(nil, g)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := g.ClientID()
			if err != nil {
				b.Fatalf("生成ID失败: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentGuard 并发基准测试：访问守卫
func BenchmarkConcurrentGuard(b *testing.B) {
	var g = setupTest(nil)
	defer cleanupTest(nil, g)

	// 预先生成多个凭证用于并发测试
	var tokens = make([]string, 100)
	for i := 0; i < 100; i++ {
		var (
			id, _    = g.ClientID()
			token, _ = g.Credential(id)
		)
		tokens[i] = token
	}

	var counter int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var idx = atomic.AddInt64(&counter, 1) % 100
			_, err := g.Guard(tokens[idx])
			if err != nil && err.Error() != "rate limit exceeded" && err.Error() != "token has expired" {
				b.Fatalf("守卫验证失败: %v", err)
			}
		}
	})
}

// ==================== 使用示例 ====================

func Example() {
	// 创建Gatekeeper实例
	var g = New(
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
			logx.LogConf{ServiceName: "example", Level: "error"},
		),
		60000, // 60秒有效期
		"fEaB5EYyHKsSbvWg",
		"node-1",
	)
	defer g.Stop()

	// 生成客户端ID
	clientID, err := g.ClientID()
	if err != nil {
		fmt.Printf("生成ID失败: %v\n", err)
		return
	}
	fmt.Printf("生成的客户端ID: %s\n", clientID)

	// 生成访问凭证
	token, err2 := g.Credential(clientID)
	if err2 != nil {
		fmt.Printf("生成凭证失败: %v\n", err2)
		return
	}
	fmt.Printf("生成的凭证: %s\n", token)

	// 使用凭证访问
	newToken, err3 := g.Guard(token)
	if err3 != nil {
		fmt.Printf("访问被拒绝: %v\n", err3)
	} else if newToken != "" {
		fmt.Printf("凭证已续期，新凭证: %s\n", newToken)
	} else {
		fmt.Println("访问成功")
	}

	// 将恶意用户加入黑名单
	err4 := g.AddToBlacklist("malicious_user", "恶意行为", 24*time.Hour)
	if err4 != nil {
		fmt.Printf("加入黑名单失败: %v\n", err4)
	}

	// 检查用户是否在黑名单中
	blacklisted, item, _ := g.IsBlacklisted("malicious_user")
	if blacklisted {
		fmt.Printf("用户在黑名单中，原因: %s, 过期时间: %v\n", item.Reason, time.UnixMilli(int64(item.Expire)))
	}

	// 设置限流
	g.SetRateLimit(10) // 每秒10次请求

	// 检查限流状态
	current, limit, resetTime, _ := g.GetRateLimitStatus(clientID)
	fmt.Printf("限流状态: 当前请求数=%d, 限制=%d, 重置时间=%v\n", current, limit, resetTime)

}
