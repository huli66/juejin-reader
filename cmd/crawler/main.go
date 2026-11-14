// Package main 负责解析参数和执行 internal 里的包
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/huli66/juejin-reader/internal/crawl"
	"github.com/huli66/juejin-reader/internal/juejin"
)

func runCrawler() error {
	bookletID := flag.String("id", "7288482920602271802", "小册的 booklet_id")
	outDir := flag.String("out", "juejin-data", "输出目录")
	interval := flag.Duration("interval", 2*time.Second, "两次请求的最小间隔")
	maxSections := flag.Int("max", 0, "最多爬几章，0 表示全部")

	cookie := os.Getenv("JUEJIN_COOKIE")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	c := juejin.New(*interval, cookie) // flag.Duration 刚执行没有数据，只有占位指针
	defer c.Close()

	if err := crawl.Run(ctx, c, *bookletID, *outDir, *maxSections); err != nil {
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
