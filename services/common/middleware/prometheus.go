package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler 返回 Prometheus 指标处理器
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

var (
	/*
		HttpRequestsTotal 记录请求总数.
		指标类型：Counter (计数器)
			特点：只增不减（除非系统重启归零）。
			用途：像汽车的里程表。主要用来计算 QPS (每秒请求数) 和 错误率。
		标签 (Labels)：method, path, status, code:
			这些标签在Grafana的看板中可以用来过滤和分析数据。如图：a-docs/images/prometheus_fiter_label.png.  code是后期加的，还没上报，所有图中没有
			动态标签："method", "path", "status", "code"。
			外部标签："service", "env"。每个服务的main.go中启动Prometheus Pusher时设置(middleware.StartMetricsPusher)。

	*/
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status", "code"}, // 核心改动：增加业务码标签
	)

	/*
		HttpRequestDuration 记录请求耗时。
		Histogram (直方图/分布图)：
			计算 P99、P95 延迟（即 99% 的用户感到多快）。
		Buckets (桶)：
			这组数字：{.005, .01, .025, ... 10} 代表了时间的“分档”。
			它是怎么工作的？：当一个请求耗时 0.03 秒时，它会落入 0.05 这个桶里（以及所有比 0.05 大的桶）。
			通过这些桶，Prometheus 就可以计算出：有多少请求是在 5ms 内完成的？有多少超过了 1 秒？这对于优化数据库查询和接口性能至关重要。
	*/
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)
)

// PrometheusMiddleware 监控中间件
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		c.Next()

		status := strconv.Itoa(c.Writer.Status()) // 将响应状态码转换为字符串
		duration := time.Since(start).Seconds()   // 计算请求耗时

		// 核心改动：从 Context 中获取业务码（biz_code），如果不存在则为 unknown
		bizCode := "unknown"
		// 从上下文中获取业务码，代码在reponse.go中
		// 返回给app的response中的code的值
		if val, exists := c.Get("biz_code"); exists {
			bizCode = fmt.Sprintf("%v", val)
		}

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status, bizCode).Inc() // 记录请求数 (含业务码)
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)    // 记录请求耗时
	}
}
