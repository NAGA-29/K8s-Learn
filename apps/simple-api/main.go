package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
)

var (
	requestCount atomic.Int64
	startTime    = time.Now()
	rdb          *redis.Client
)

func main() {
	// Redis は /api/count でのみ使用する。接続は遅延確立されるため、
	// Redis が存在しない環境（Step08/09 など）でも起動とヘルスチェックには影響しない
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "redis"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	rdb = redis.NewClient(&redis.Options{
		Addr:        redisHost + ":" + redisPort,
		DialTimeout: 2 * time.Second,
	})

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(countRequests)

	e.GET("/health", healthHandler)
	e.GET("/api/message", messageHandler)
	e.GET("/api/count", countHandler)
	e.GET("/metrics", metricsHandler)

	e.Logger.Fatal(e.Start(":8080"))
}

func countRequests(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestCount.Add(1)
		return next(c)
	}
}

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func messageHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message":   "Hello from simple-api!",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func countHandler(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	count, err := rdb.Incr(ctx, "visit_count").Result()
	if err != nil {
		c.Logger().Errorf("redis error: %v", err)
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "redis unavailable: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"count":     count,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func metricsHandler(c echo.Context) error {
	uptime := time.Since(startTime).Seconds()
	body := fmt.Sprintf(`# HELP request_count_total Total number of requests received.
# TYPE request_count_total counter
request_count_total %d

# HELP goroutine_count Current number of goroutines.
# TYPE goroutine_count gauge
goroutine_count %d

# HELP uptime_seconds Uptime in seconds.
# TYPE uptime_seconds gauge
uptime_seconds %.2f
`, requestCount.Load(), runtime.NumGoroutine(), uptime)

	return c.String(http.StatusOK, body)
}
