package juejin

import "encoding/json"

// ========== struct ==========

// Resp 返回结果结构体
type Resp[T any] struct {
	ErrNo  int    `json:"err_no"`
	ErrMsg string `json:"err_msg"`
	Data   T      `json:"data"`
	// 评论接口字段
	Cursor  string `json:"cursor"`
	HasMore bool   `json:"has_more"`
	Count   int    `json:"count"`
}

// BookletData 是书集合
type BookletData struct {
	Booklet struct {
		BaseInfo struct {
			Title        string `json:"title"`
			SectionCount int    `json:"section_count"`
		} `json:"base_info"`
	} `json:"booklet"`
	Sections []SectionBrief `json:"sections"`
}

// SectionBrief 是章节 id 和评论数量
type SectionBrief struct {
	SectionID    string `json:"section_id"`
	Title        string `json:"title"`
	CommentCount int    `json:"comment_count"`
}

// SectionData 是章节结构体
type SectionData struct {
	Section struct {
		Title        string `json:"title"`
		MarkdownShow string `json:"markdown_show"`
	} `json:"section"`
}

// Comment 是评论结构体
type Comment struct {
	CommentInfo struct {
		CommentID      string `json:"comment_id"`
		CommentContent string `json:"comment_content"`
		Ctime          int64  `json:"ctime"`
		DiggCount      int    `json:"digg_count"`
		ReplyCount     int    `json:"reply_count"`
	} `json:"comment_info"`
	UserInfo struct {
		UserName string `json:"user_name"`
	} `json:"user_info"`
	ReplyInfos []json.RawMessage `json:"reply_infos"`
}
