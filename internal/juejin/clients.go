package juejin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"
)

// Client 是掘金 API 的客户端
type Client struct {
	baseURL string
	cookie  string
	backoff time.Duration
	http    *http.Client
	ticker  *time.Ticker
}

// Close 是关闭 Client 的 ticker
func (c *Client) Close() {
	c.ticker.Stop()
}

// New 创建一个掘金 API 客户端
func New(interval time.Duration, cookie string) *Client {
	return &Client{
		baseURL: "https://api.juejin.cn",
		backoff: 500 * time.Millisecond,
		http:    &http.Client{Timeout: 15 * time.Second},
		ticker:  time.NewTicker(interval),
		cookie:  cookie,
	}
}

const maxAttempts = 3

func (c *Client) post[T any](ctx context.Context, path string, body any, out *Resp[T]) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("post %s: encode: %w", path, err)
	}

	for attempt := 1; ; attempt++ {
		err = c.postOnce(ctx, path, buf, out)
		if err == nil {
			return nil
		}
		if !isRetryable(ctx, err) {
			return err
		}
		if attempt == maxAttempts {
			return fmt.Errorf("after %d attempts: %w", maxAttempts, err)
		}

		// 退避：500ms 1s 每次翻倍，再加一点随机抖动
		// 左移 n 位是乘以 2 的 n 次方
		delay := c.backoff << (attempt - 1)
		delay += rand.N(delay/2 + 1)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
}

func (c *Client) postOnce[T any](ctx context.Context, path string, buf []byte, out *Resp[T]) error {
	// 限速：等下一个 tick，同时能被 Ctrl+C 打断
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ticker.C:
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path+"?aid=2608", bytes.NewReader(buf))
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
		return &HTTPError{Path: path, StatusCode: resp.StatusCode}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("post %s: decode: %w %w", path, errDecode, err)
	}

	// HTTP 200 只代表请求成功，不代表业务成功
	if out.ErrNo != 0 {
		return &APIError{Path: path, ErrNo: out.ErrNo, ErrMsg: out.ErrMsg}
	}
	return nil
}

func isRetryable(ctx context.Context, err error) bool {
	// 用户取消了或者 ctx 超时（不是 http 超时），不重试
	if ctx.Err() != nil {
		return false
	}
	var httpErr *HTTPError        // 惯例错误都用指针接收者
	if errors.As(err, &httpErr) { // As 也匹配指针接收者
		return httpErr.StatusCode >= 500 || httpErr.StatusCode == http.StatusTooManyRequests
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return false
	}
	if errors.Is(err, errDecode) {
		return false
	}
	// 剩下的都是网络层错误
	return true
}
