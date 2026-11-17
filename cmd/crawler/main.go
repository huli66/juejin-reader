// Package main 负责解析参数和执行 internal 里的包
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/huli66/juejin-reader/internal/crawl"
	"github.com/huli66/juejin-reader/internal/juejin"
)

func runCrawler() error {
	bookletID := flag.String("id", "7288482920602271802", "小册的 booklet_id")
	// 第三个参数是用途，-h 会输出该内容，反引号包裹部分会替代输出中的类型
	outDir := flag.String("out", "juejin-data", "输出到`目录`")
	interval := flag.Duration("interval", 2*time.Second, "两次请求的最小间隔")
	maxSections := flag.Int("max", 0, "最多爬几章，0 表示全部")
	concurrency := flag.Int("concurrency", 1, "同时爬取几章")

	cookie := os.Getenv("JUEJIN_COOKIE")
	flag.Parse()
	// 校验必须写在 Parse 后面，否则会使用默认值来判断
	if *interval <= 0 {
		return fmt.Errorf("-interval 必须大于 0，当前是 %v", *interval)
	}
	if *concurrency <= 0 {
		return fmt.Errorf("-concurrency 必须大于 0，当前是 %v", *concurrency)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	c := juejin.New(*interval, cookie) // flag.Duration 刚执行没有数据，只有占位指针
	defer c.Close()

	if err := crawl.Run(ctx, c, crawl.Options{
		BookletID:   *bookletID,
		OutDir:      *outDir,
		MaxSections: *maxSections,
		Concurrency: *concurrency,
	}); err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("已中断，下次会从未完成的章节继续")
			return nil
		}
		return err
	}
	log.Println("完成")
	return nil
}

func main() {
	err := runCrawler()
	if err != nil {
		log.Fatal(err)
	}
}
