package juejin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// newTestClient 起一个假的掘金服务，返回指向它的 Client
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper() // 断言失败时，报错行号指向调用方而不是这里

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close) // 测试结束时自动关闭，类似 afterEach

	c := New(time.Millisecond, "test-cookie")
	t.Cleanup(c.Close)
	c.baseURL = srv.URL // 同包测试可以直接改未导出字段
	c.backoff = time.Millisecond

	return c
}

func TestGetBooklet(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// handler 跑在服务端 goroutine 里，只能用 t.Errorf 不能用 t.Fatalf
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}

		if r.URL.Path != "/booklet_api/v1/booklet/get" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("aid"); got != "2608" {
			t.Errorf("aid = %q want 2608", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}
		if got := r.Header.Get("Cookie"); got != "test-cookie" {
			t.Errorf("Cookie = %q", got)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body["booklet_id"] != "123" {
			t.Errorf("booklet_id = %q, want 123", body["booklet_id"])
		}

		_, _ = fmt.Fprint(w, `{
			"err_no": 0,
			"data": {
				"booklet": {
					"base_info": {
						"title": "测试小测",
						"section_count": 2
					}
				},
				"sections": [
					{"section_id": "s1", "title": "第一章"},
					{"section_id": "s2", "title": "第二章"}
				]
			}
		}`)
	})

	got, err := c.GetBooklet(context.Background(), "123")
	if err != nil {
		t.Fatalf("GetBooklet() error = %v", err)
	}
	if got.Booklet.BaseInfo.Title != "测试小测" {
		t.Errorf("title = %q", got.Booklet.BaseInfo.Title)
	}
	if len(got.Sections) != 2 {
		t.Fatalf("len(sections) = %d, want 2", len(got.Sections))
	}
	if got.Sections[1].SectionID != "s2" {
		t.Errorf("sections[1].id = %q, want s2", got.Sections[1].SectionID)
	}
}

func TestPostErrors(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantMsg string // 错误信息里应该包含的内容
	}{
		{
			name: "HTTP 500",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantMsg: "http 500",
		},
		{
			name: "业务错误",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprint(w, `{"err_no": 403, "err_msg": "未购买"}`)
			},
			wantMsg: "未购买",
		},
		{
			name: "响应不是 JSON",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprint(w, `<html>`)
			},
			wantMsg: "decode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, tt.handler)

			_, err := c.GetSection(context.Background(), "s1")
			if err == nil {
				t.Fatal("want error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("error = %q, want contains %q", err, tt.wantMsg)
			}
		})
	}
}

func TestPostCanceled(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"err_no": 0}`)
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 发请求之前就取消

	_, err := c.GetSection(ctx, "s1")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

func TestGetAllCommentsPagination(t *testing.T) {
	var calls atomic.Int32 // handler 和测试在不同 goroutine，计数要用原子类型

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)

		var body struct {
			ItemID string `json:"item_id"`
			Cursor string `json:"cursor"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body.ItemID != "s1" {
			t.Errorf("item_id = %q, want s1", body.ItemID)
		}

		switch body.Cursor {
		case "0": // 第一页
			_, _ = fmt.Fprint(w, `{
				"err_no": 0, "cursor": "2", "has_more": true,
				"data": [
					{"comment_info": {"comment_id": "c1"}},
					{"comment_info": {"comment_id": "c2"}}
				]
			}`)
		case "2": // 第二页，也是最后一页
			_, _ = fmt.Fprint(w, `{
				"err_no": 0, "cursor": "3", "has_more": false,
				"data": [{"comment_info": {"comment_id": "c3"}}]
			}`)
		default:
			t.Errorf("unexpected cursor %q", body.Cursor)
		}
	})

	got, err := c.GetAllComments(context.Background(), "s1")
	if err != nil {
		t.Fatalf("GetAllComments() error = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len(comments) = %d, want 3", len(got))
	}
	if got[2].CommentInfo.CommentID != "c3" {
		t.Errorf("comments[2].id = %q, want c3", got[2].CommentInfo.CommentID)
	}
	if n := calls.Load(); n != 2 {
		t.Errorf("请求次数 = %d, want 2", n)
	}
}

func TestPostRetry(t *testing.T) {
	const okBody = `{"err_no": 0, "data": {"section": {"title": "ok"}}}`

	tests := []struct {
		name      string
		respond   func(n int32, w http.ResponseWriter) // n 是第几次请求，从 1 开始
		wantCalls int32
		check     func(t *testing.T, err error)
	}{
		{
			name: "前两次 500，第三次成功",
			respond: func(n int32, w http.ResponseWriter) {
				if n < 3 {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				_, _ = fmt.Fprint(w, okBody)
			},
			wantCalls: 3,
			check: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("error = %v, want nil", err)
				}
			},
		},
		{
			name: "一直 500，重试用完",
			respond: func(_ int32, w http.ResponseWriter) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantCalls: 3,
			check: func(t *testing.T, err error) {
				var httpErr *HTTPError
				if !errors.As(err, &httpErr) {
					t.Fatalf("error = %v, want *HTTPError", err)
				}
				if httpErr.StatusCode != http.StatusInternalServerError {
					t.Errorf("status = %d, want 500", httpErr.StatusCode)
				}
			},
		},
		{
			name: "429 后成功",
			respond: func(n int32, w http.ResponseWriter) {
				if n == 1 {
					w.WriteHeader(http.StatusTooManyRequests)
					return
				}
				_, _ = fmt.Fprint(w, okBody)
			},
			wantCalls: 2,
			check: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("error = %v, want nil", err)
				}
			},
		},
		{
			name: "404 不重试",
			respond: func(_ int32, w http.ResponseWriter) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantCalls: 1,
			check: func(t *testing.T, err error) {
				var httpErr *HTTPError
				if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusNotFound {
					t.Errorf("error = %v, want *HTTPError 404", err)
				}
			},
		},
		{
			name: "业务错误不重试",
			respond: func(_ int32, w http.ResponseWriter) {
				_, _ = fmt.Fprint(w, `{"err_no": 403, "err_msg": "未购买"}`)
			},
			wantCalls: 1,
			check: func(t *testing.T, err error) {
				var apiErr *APIError
				if !errors.As(err, &apiErr) || apiErr.ErrNo != 403 {
					t.Errorf("error = %v, want *APIError 403", err)
				}
			},
		},
		{
			name: "响应不是 JSON 不重试",
			respond: func(_ int32, w http.ResponseWriter) {
				_, _ = fmt.Fprint(w, `<html>`)
			},
			wantCalls: 1,
			check: func(t *testing.T, err error) {
				if !errors.Is(err, errDecode) {
					t.Errorf("error = %v, want errDecode", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				tt.respond(calls.Add(1), w)
			})

			_, err := c.GetSection(context.Background(), "s1")

			tt.check(t, err)
			if n := calls.Load(); n != tt.wantCalls {
				t.Errorf("请求次数 = %d, want %d", n, tt.wantCalls)
			}
		})
	}
}

func TestPostCanceledDuringBackoff(t *testing.T) {
	var calls atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	})
	c.backoff = 10 * time.Second // 故意设得很长，保证取消时正卡在退避等待里

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	time.AfterFunc(50*time.Millisecond, cancel) // 50ms 后取消

	_, err := c.GetSection(ctx, "s1")

	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
	if n := calls.Load(); n != 1 {
		t.Errorf("请求次数 = %d, want 1", n)
	}
}
