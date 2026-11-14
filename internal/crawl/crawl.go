// Package crawl 是爬虫本体
package crawl

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/huli66/juejin-reader/internal/juejin"
)

// Run 是执行爬虫，注释必须以导出名开头
func Run(ctx context.Context, c *juejin.Client, bookletID, outDir string, maxSections int) error {
	book, err := c.GetBooklet(ctx, bookletID)
	if err != nil {
		return err
	}
	title := book.Booklet.BaseInfo.Title
	dir := filepath.Join(outDir, safeName(title))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	log.Printf("课程《%s》共 %d 章", title, len(book.Sections))

	for i, s := range book.Sections {
		if maxSections > 0 && i >= maxSections {
			break
		}
		prefix := filepath.Join(dir, fmt.Sprintf("%02d_%s", i+1, safeName(s.Title)))
		mdPath, commentPath := prefix+".md", prefix+".comments.json"

		// 断点续爬：两个文件都在就跳过
		if fileExists(mdPath) && fileExists(commentPath) {
			log.Printf("[%d] 已存在，跳过: %s", i+1, s.Title)
			continue
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
	}
	return nil
}
