package juejin

import (
	"errors"
	"fmt"
)

// HTTPError 表示 HTTP 状态码不是 200
type HTTPError struct {
	Path       string
	StatusCode int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("post %s: http %d", e.Path, e.StatusCode)
}

// APIError 表示 HTTP 成功，但是业务失败
type APIError struct {
	Path   string
	ErrNo  int
	ErrMsg string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("post %s: err_no=%d %s", e.Path, e.ErrNo, e.ErrMsg)
}

// errDecode 标记响应解析失败，用 errors.Is 判断
var errDecode = errors.New("decode response")
