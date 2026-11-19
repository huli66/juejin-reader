package crawl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/huli66/juejin-reader/internal/juejin"
)

// Store 表示一本书的章节存储，不同 worker 可能同时调用它处理不同章节
type Store interface {
	HasSection(
		ctx context.Context,
		index int,
		brief juejin.SectionBrief,
	) (bool, error)

	SaveSection(
		ctx context.Context,
		index int,
		brief juejin.SectionBrief,
		section juejin.SectionData,
		comments []juejin.Comment,
	) error
}

// FileStore 保存在文件目录
type FileStore struct {
	dir string
}

// NewFileStore 创建一本书对应的文件存储
func NewFileStore(outDir, bookTitle string) (*FileStore, error) {
	dir := filepath.Join(outDir, safeName(bookTitle))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &FileStore{dir: dir}, nil
}

func (f *FileStore) paths(index int, brief juejin.SectionBrief) (string, string) {
	prefix := filepath.Join(
		f.dir,
		fmt.Sprintf("%02d_%s", index+1, safeName(brief.Title)),
	)
	return prefix + ".md", prefix + ".comments.json"
}

// HasSection 判断是否已存在
func (f *FileStore) HasSection(
	ctx context.Context,
	index int,
	brief juejin.SectionBrief,
) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	mdPath, commentPath := f.paths(index, brief)
	return fileExists(mdPath) && fileExists(commentPath), nil
}

// SaveSection 保存章节正文和评论
func (f *FileStore) SaveSection(
	ctx context.Context,
	index int,
	brief juejin.SectionBrief,
	section juejin.SectionData,
	comments []juejin.Comment,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	mdPath, commentPath := f.paths(index, brief)
	md := "# " + section.Section.Title + "\n\n" + section.Section.MarkdownShow

	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		return err
	}
	return saveJSON(commentPath, comments)
}
