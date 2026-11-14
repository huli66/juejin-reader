package juejin

import "context"

// GetBooklet 获取书
func (c *Client) GetBooklet(ctx context.Context, id string) (BookletData, error) {

	var r Resp[BookletData]
	err := c.post(ctx, "/booklet_api/v1/booklet/get", map[string]string{"booklet_id": id}, &r)

	return r.Data, err
}

// GetSection 获取章节
func (c *Client) GetSection(ctx context.Context, id string) (SectionData, error) {
	var r Resp[SectionData]
	err := c.post(ctx, "/booklet_api/v1/section/get", map[string]string{"section_id": id}, &r)
	return r.Data, err
}

// GetAllComments 获取全部评论
func (c *Client) GetAllComments(ctx context.Context, sectionID string) ([]Comment, error) {
	var all []Comment
	cursor := "0"

	for {
		var r Resp[[]Comment]
		body := map[string]any{
			"item_id":     sectionID,
			"item_type":   13,
			"cursor":      cursor,
			"limit":       20,
			"sort":        0,
			"client_type": 2608,
		}

		if err := c.post(ctx, "/interact_api/v1/comment/list", body, &r); err != nil {
			return all, err
		}
		all = append(all, r.Data...)
		if !r.HasMore || r.Cursor == cursor {
			return all, nil
		}
		cursor = r.Cursor
	}
}
