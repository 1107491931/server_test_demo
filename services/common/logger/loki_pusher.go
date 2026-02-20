package logger

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/**
文件的作用：将日志推送到Grafana Loki.
*/

// LokiCore 实现 zapcore.Core 接口，用于将日志(zapcore.Core)推送到Grafana Loki.
type LokiCore struct {
	zapcore.LevelEnabler
	encoder zapcore.Encoder
	pusher  *lokiPusher
}

func NewLokiCore(cfg LokiConfig, level zapcore.Level) zapcore.Core {
	if !cfg.Enabled || cfg.URL == "" {
		return zapcore.NewNopCore()
	}

	// 使用 JSON 编码器，这样推送到 Loki 的日志就是 JSON 格式
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	encoder := zapcore.NewJSONEncoder(encCfg)

	pusher := newLokiPusher(cfg)

	return &LokiCore{
		LevelEnabler: level,
		encoder:      encoder,
		pusher:       pusher,
	}
}

func (c *LokiCore) With(fields []zapcore.Field) zapcore.Core {
	clone := c.encoder.Clone()
	for _, f := range fields {
		f.AddTo(clone)
	}
	return &LokiCore{
		LevelEnabler: c.LevelEnabler,
		encoder:      clone,
		pusher:       c.pusher,
	}
}

func (c *LokiCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *LokiCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	buf, err := c.encoder.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}

	line := buf.String()
	buf.Free()

	c.pusher.push(ent.Time, line)
	return nil
}

func (c *LokiCore) Sync() error {
	return nil
}

// lokiPusher 处理异步批量推送
type lokiPusher struct {
	cfg    LokiConfig
	client *http.Client
	ch     chan logEntry
}

type logEntry struct {
	ts   time.Time
	line string
}

func newLokiPusher(cfg LokiConfig) *lokiPusher {
	p := &lokiPusher{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
		ch:     make(chan logEntry, 1000),
	}
	go p.run()
	return p
}

func (p *lokiPusher) push(ts time.Time, line string) {
	select {
	case p.ch <- logEntry{ts: ts, line: line}:
	default:
		// 如果队列满了，丢弃日志以防止阻塞主流程
	}
}

func (p *lokiPusher) run() {
	ticker := time.NewTicker(2 * time.Second)
	var batch []logEntry
	maxBatch := 500

	for {
		select {
		case entry := <-p.ch:
			batch = append(batch, entry)
			if len(batch) >= maxBatch {
				p.send(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				p.send(batch)
				batch = nil
			}
		}
	}
}

type lokiPushRequest struct {
	Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"`
}

func (p *lokiPusher) send(batch []logEntry) {
	reqBody := lokiPushRequest{
		Streams: []lokiStream{
			{
				Stream: p.cfg.Labels,
				Values: make([][2]string, len(batch)),
			},
		},
	}

	for i, entry := range batch {
		reqBody.Streams[0].Values[i] = [2]string{
			strconv.FormatInt(entry.ts.UnixNano(), 10),
			entry.line,
		}
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return
	}

	// 压缩
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return
	}
	gw.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url := p.cfg.URL
	// 确保 URL 包含完整路径
	// 如果用户只提供了域名，补全路径
	if !bytes.Contains([]byte(url), []byte("/loki/api/v1/push")) && !bytes.Contains([]byte(url), []byte("/api/prom/push")) {
		if url[len(url)-1] == '/' {
			url += "loki/api/v1/push"
		} else {
			url += "/loki/api/v1/push"
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("User-Agent", "we-circle-loki-pusher/1.0")

	if p.cfg.UserID != "" && p.cfg.Token != "" {
		// 统一使用 Bearer 格式解决 Grafana Cloud EOF 问题
		auth := fmt.Sprintf("%s:%s", p.cfg.UserID, p.cfg.Token)
		req.Header.Set("Authorization", "Bearer "+auth)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		fmt.Printf("[Loki] Failed to send logs: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var resBody bytes.Buffer
		resBody.ReadFrom(resp.Body)
		fmt.Printf("[Loki] Unexpected status %d: %s\n", resp.StatusCode, resBody.String())
	}
}
