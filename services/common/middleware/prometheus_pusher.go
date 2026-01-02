package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"go.uber.org/zap"
)

// MetricsPusherConfig 包含远程写入 Prometheus 所需的配置
type MetricsPusherConfig struct {
	Enabled  bool
	URL      string
	UserID   string
	Token    string
	Interval time.Duration
	Labels   map[string]string
}

// StartMetricsPusher 启动一个后台 goroutine 周期性推送指标
func StartMetricsPusher(ctx context.Context, cfg MetricsPusherConfig, logger *zap.Logger) {
	if !cfg.Enabled || cfg.URL == "" {
		return
	}

	if cfg.Interval == 0 {
		cfg.Interval = 15 * time.Second
	}

	logger.Info("Starting Prometheus Metrics Pusher",
		zap.String("url", cfg.URL),
		zap.String("user_id", cfg.UserID),
		zap.Duration("interval", cfg.Interval),
	)

	go func() {
		ticker := time.NewTicker(cfg.Interval) // 每15秒触发一次
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := pushMetrics(cfg); err != nil {
					logger.Error("Failed to push metrics to Prometheus", zap.Error(err))
				}
			}
		}
	}()
}

func pushMetrics(cfg MetricsPusherConfig) error {
	// 1. 从默认收集器获取所有当前指标
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return fmt.Errorf("gather metrics: %w", err)
	}

	now := time.Now().UnixNano() / int64(time.Millisecond)
	var timeseries []prompb.TimeSeries

	// 2. 将 Prometheus 指标转换为 Remote Write 格式 (TimeSeries)
	for _, mf := range metricFamilies {
		for _, m := range mf.Metric {
			var labels []prompb.Label

			// 添加指标名作为 __name__ 标签
			labels = append(labels, prompb.Label{
				Name:  model.MetricNameLabel,
				Value: mf.GetName(),
			})

			// 添加指标自带的标签
			for _, l := range m.Label {
				labels = append(labels, prompb.Label{
					Name:  l.GetName(),
					Value: l.GetValue(),
				})
			}

			// 添加全局配置的标签 (如 service, env)
			for k, v := range cfg.Labels {
				labels = append(labels, prompb.Label{
					Name:  k,
					Value: v,
				})
			}

			// 处理不同类型的指标，拆分为多个 TimeSeries (针对 Histogram)
			if m.Gauge != nil {
				timeseries = append(timeseries, createTimeSeries(mf.GetName(), labels, m.Gauge.GetValue(), now))
			} else if m.Counter != nil {
				timeseries = append(timeseries, createTimeSeries(mf.GetName(), labels, m.Counter.GetValue(), now))
			} else if m.Untyped != nil {
				timeseries = append(timeseries, createTimeSeries(mf.GetName(), labels, m.Untyped.GetValue(), now))
			} else if m.Histogram != nil {
				// 导出三个子指标：_sum, _count, _bucket
				timeseries = append(timeseries, createTimeSeries(mf.GetName()+"_sum", labels, m.Histogram.GetSampleSum(), now))
				timeseries = append(timeseries, createTimeSeries(mf.GetName()+"_count", labels, float64(m.Histogram.GetSampleCount()), now))

				// 导出 buckets
				for _, b := range m.Histogram.Bucket {
					bucketLabels := append([]prompb.Label{}, labels...)
					bucketLabels = append(bucketLabels, prompb.Label{
						Name:  "le",
						Value: fmt.Sprintf("%g", b.GetUpperBound()),
					})
					timeseries = append(timeseries, createTimeSeries(mf.GetName()+"_bucket", bucketLabels, float64(b.GetCumulativeCount()), now))
				}
			} else if m.Summary != nil {
				timeseries = append(timeseries, createTimeSeries(mf.GetName()+"_sum", labels, m.Summary.GetSampleSum(), now))
				timeseries = append(timeseries, createTimeSeries(mf.GetName()+"_count", labels, float64(m.Summary.GetSampleCount()), now))
			}
		}
	}

	if len(timeseries) == 0 {
		return nil
	}

	// 3. 序列化为 Protobuf
	req := &prompb.WriteRequest{Timeseries: timeseries}
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal protobuf: %w", err)
	}

	// 4. 使用 Snappy 压缩 (Prometheus Remote Write 强制要求)
	compressed := snappy.Encode(nil, data)

	// 5. 发送 HTTP POST
	httpReq, err := http.NewRequest("POST", cfg.URL, bytes.NewReader(compressed))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Encoding", "snappy")
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	httpReq.Header.Set("User-Agent", "we-circle-metrics-pusher/1.0")

	if cfg.UserID != "" && cfg.Token != "" {
		// 针对 Grafana Cloud 的 Access Policy Token，使用 Bearer ID:Token 格式更稳定
		auth := fmt.Sprintf("%s:%s", cfg.UserID, cfg.Token)
		httpReq.Header.Set("Authorization", "Bearer "+auth)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// 调试日志：确认发送成功
	fmt.Printf("Successfully pushed %d timeseries to Prometheus\n", len(timeseries))

	return nil
}

func createTimeSeries(name string, labels []prompb.Label, value float64, timestamp int64) prompb.TimeSeries {
	// 深度复制标签，防止冲突
	newLabels := make([]prompb.Label, 0, len(labels))
	for _, l := range labels {
		if l.Name == model.MetricNameLabel {
			newLabels = append(newLabels, prompb.Label{Name: l.Name, Value: name})
			continue
		}
		newLabels = append(newLabels, l)
	}

	return prompb.TimeSeries{
		Labels: newLabels,
		Samples: []prompb.Sample{
			{Value: value, Timestamp: timestamp},
		},
	}
}
