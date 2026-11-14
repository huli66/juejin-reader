package juejin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 是掘金 API 的客户端
type Client struct {
	baseURL string
	http    *http.Client
	ticker  *time.Ticker
	cookie  string
}

// Close 是关闭 Client 的 ticker
func (c *Client) Close() {
	c.ticker.Stop()
}

// New 创建一个掘金 API 客户端
func New(interval time.Duration, cookie string) *Client {
	return &Client{
		baseURL: "https://api.juejin.cn",
		http:    &http.Client{Timeout: 15 * time.Second},
		ticker:  time.NewTicker(interval),
		cookie:  cookie,
	}
}

func (c *Client) post[T any](ctx context.Context, path string, body any, out *Resp[T]) error {
	// 限速：等下一个 tick，同时能被 Ctrl+C 打断
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ticker.C:
	}

	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path+"?aid=2608", bytes.NewReader((buf)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if c.cookie != "" {
		req.Header.Set("Cookie", c.cookie)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("post %s: %w", path, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("post %s: http %d", path, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("post %s: decode: %w", path, err)
	}

	// HTTP 200 只代表请求成功，不代表业务成功
	if out.ErrNo != 0 {
		return fmt.Errorf("post %s: err_no=%d %s", path, out.ErrNo, out.ErrMsg)
	}
	return nil
}
