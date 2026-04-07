package main

import (
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	requestCount atomic.Int64
	startTime    = time.Now()
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(countRequests)

	e.GET("/health", healthHandler)
	e.GET("/api/message", messageHandler)
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
