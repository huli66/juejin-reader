// Package crawl 是爬虫本体
package crawl

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/huli66/juejin-reader/internal/juejin"
)

func crawlSection(ctx context.Context, c *juejin.Client, dir string, i int, s juejin.SectionBrief) error {
	prefix := filepath.Join(dir, fmt.Sprintf("%02d_%s", i+1, safeName(s.Title)))
	mdPath, commentPath := prefix+".md", prefix+".comments.json"

	// 断点续爬：两个文件都在就跳过
	if fileExists(mdPath) && fileExists(commentPath) {
		log.Printf("[%d] 已存在，跳过: %s", i+1, s.Title)
		return nil
	}

	sec, err := c.GetSection(ctx, s.SectionID)
	if err != nil {
		return err
	}
	md := "# " + sec.Section.Title + "\n\n" + sec.Section.MarkdownShow
	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		return err
	}

	comments, err := c.GetAllComments(ctx, s.SectionID)
	if err != nil {
		return err
	}
	if err := saveJSON(commentPath, comments); err != nil {
		return err
	}
	log.Printf("[%d] %s: 正文 %d 字， 评论 %d 条", i+1, s.Title, len([]rune(sec.Section.MarkdownShow)), len(comments))
	return nil
}

type job struct {
	index   int
	section juejin.SectionBrief
}

// Options 是执行爬虫的参数集合
type Options struct {
	BookletID   string
	OutDir      string
	MaxSections int
	Concurrency int
}

// Run 是执行爬虫，注释必须以导出名开头
func Run(ctx context.Context, c *juejin.Client, opt Options) error {
	if opt.Concurrency <= 0 {
		return fmt.Errorf(
			"concurrency 必须大于 0, 当前是 %d",
			opt.Concurrency,
		)
	}
	book, err := c.GetBooklet(ctx, opt.BookletID)
	if err != nil {
		return err
	}
	title := book.Booklet.BaseInfo.Title
	dir := filepath.Join(opt.OutDir, safeName(title))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	log.Printf("课程《%s》共 %d 章", title, len(book.Sections))

	jobs := make(chan job)
	g, workerCtx := errgroup.WithContext(ctx)

	// 启动固定数量的 worker
	for w := 0; w < opt.Concurrency; w++ {
		g.Go(func() error {
			for {
				select {
				case <-workerCtx.Done():
					return workerCtx.Err()
				case task, ok := <-jobs:
					if !ok {
						return nil
					}
					if err := workerCtx.Err(); err != nil {
						return err
					}
					err := crawlSection(
						workerCtx,
						c,
						dir,
						task.index,
						task.section,
					)
					if err != nil {
						return fmt.Errorf(
							"爬取第 %d 章 %q: %w",
							task.index+1,
							task.section.Title,
							err,
						)
					}
				}
			}
		})
	}

	// 生产者负责发送任务，并在退出时关闭通道
	g.Go(func() error {
		defer close(jobs)
		for i, s := range book.Sections {
			if opt.MaxSections > 0 && i >= opt.MaxSections {
				break
			}
			select {
			case <-workerCtx.Done():
				return workerCtx.Err()
			case jobs <- job{index: i, section: s}:

			}
		}
		return nil
	})

	// 等待生产者和全部 worker 退出
	if err := g.Wait(); err != nil {
		return err
	}
	return ctx.Err()

	// for i, s := range book.Sections {
	// 	if opt.MaxSections > 0 && i >= opt.MaxSections {
	// 		break
	// 	}
	// 	err := crawlSection(ctx, c, dir, i, s)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// return nil
}
