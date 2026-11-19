package crawl

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // 注册名为 sqlite 的 database/sql 驱动。

	"github.com/huli66/juejin-reader/internal/juejin"
)

const sqliteSchema = `
CREATE TABLE IF NOT EXISTS books (
    booklet_id TEXT PRIMARY KEY NOT NULL,
    title      TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sections (
    booklet_id    TEXT NOT NULL,
    section_id    TEXT NOT NULL,
    section_index INTEGER NOT NULL,
    title         TEXT NOT NULL,
    markdown      TEXT NOT NULL,
    comments_json TEXT NOT NULL,
    PRIMARY KEY (booklet_id, section_id),
    FOREIGN KEY (booklet_id) REFERENCES books(booklet_id)
);
`

// SQLiteStore 把一本书的章节保存在 SQLite 中。
type SQLiteStore struct {
	db        *sql.DB
	bookletID string
}

// NewSQLiteStore 打开数据库、初始化表，并登记当前书籍。
func NewSQLiteStore(
	ctx context.Context,
	dbPath string,
	bookletID string,
	bookTitle string,
) (*SQLiteStore, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return nil, err
	}

	// 使用 file URI，正确处理路径中的空格、? 等字符。
	// 每个新连接都启用外键，并设置 5 秒的锁等待时间。
	query := url.Values{}
	query.Add("_pragma", "foreign_keys(1)")
	query.Add("_pragma", "busy_timeout(5000)")

	dsn := (&url.URL{
		Scheme:   "file",
		Path:     filepath.ToSlash(absPath),
		RawQuery: query.Encode(),
	}).String()

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// 多个 worker 共用一个数据库连接，数据库操作排队执行。
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// ExecContext 会实际建立连接；sql.Open 本身可能尚未连接。
	if _, err := db.ExecContext(ctx, sqliteSchema); err != nil {
		_ = db.Close()
		return nil, sqliteOpError(ctx, "初始化数据库", err)
	}

	_, err = db.ExecContext(ctx, `
        INSERT INTO books (booklet_id, title)
        VALUES (?, ?)
        ON CONFLICT(booklet_id) DO UPDATE SET
            title = excluded.title
    `, bookletID, bookTitle)
	if err != nil {
		_ = db.Close()
		return nil, sqliteOpError(ctx, "保存书籍信息", err)
	}

	return &SQLiteStore{
		db:        db,
		bookletID: bookletID,
	}, nil
}

// HasSection 按书籍 ID 和章节 ID 判断是否已经保存。
func (s *SQLiteStore) HasSection(
	ctx context.Context,
	_ int,
	brief juejin.SectionBrief,
) (bool, error) {
	var exists bool

	err := s.db.QueryRowContext(ctx, `
        SELECT EXISTS(
            SELECT 1 FROM sections
            WHERE booklet_id = ? AND section_id = ?
        )
    `, s.bookletID, brief.SectionID).Scan(&exists)
	if err != nil {
		return false, sqliteOpError(ctx, "查询章节", err)
	}

	return exists, nil
}

// SaveSection 在同一条 SQL 中保存正文和评论。
func (s *SQLiteStore) SaveSection(
	ctx context.Context,
	index int,
	brief juejin.SectionBrief,
	section juejin.SectionData,
	comments []juejin.Comment,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	commentsJSON, err := json.Marshal(comments)
	if err != nil {
		return fmt.Errorf("编码评论: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
        INSERT INTO sections (
            booklet_id, section_id, section_index,
            title, markdown, comments_json
        )
        VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT(booklet_id, section_id) DO UPDATE SET
            section_index = excluded.section_index,
            title = excluded.title,
            markdown = excluded.markdown,
            comments_json = excluded.comments_json
    `,
		s.bookletID,
		brief.SectionID,
		index,
		section.Section.Title,
		section.Section.MarkdownShow,
		string(commentsJSON),
	)
	if err != nil {
		return sqliteOpError(ctx, "保存章节", err)
	}

	return nil
}

// Close 释放数据库连接。
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// 保留 context 取消语义，便于上层用 errors.Is 判断取消。
func sqliteOpError(
	ctx context.Context,
	operation string,
	err error,
) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	return fmt.Errorf("%s: %w", operation, err)
}
