功能里程碑 → 对应的学习点

这样基础知识就在用到的时候顺带学了：

里程碑	用到时顺带学
M1 迁移爬虫到新结构，make lint 全部通过	包设计、导出规则；lint 报出来的问题会逼你学 error 处理
M2 爬虫支持并发、重试、Ctrl+C 退出	errgroup、worker pool、context、signal.NotifyContext、-race
M3 抽出 Store 接口，先实现文件版，再实现 SQLite 版	接口（这次能真正弄懂为什么用接口）、database/sql
M4 HTTP API：书籍列表、章节、搜索	net/http、中间件、JSON、httptest、%w / errors.As
M5 React 前端调用 API	CORS；前端本来就是你的强项
M6 用 embed 把前端打进二进制，交叉编译，CI 自动发布	embed、go build 交叉编译、GitHub Actions
穿插	发现重复代码时学泛型，比如通用的 retry[T]、分页结果 Page[T]
