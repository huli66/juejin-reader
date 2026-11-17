# 阶段一：在 Linux 环境中编译 Go 程序。
FROM golang:1.27.4-alpine3.24 AS builder

WORKDIR /src
# .dockerignore 只放行 go.mod、go.sum（如果存在）和 Go 源码目录。
COPY . .
# CPU 架构跟随构建阶段；无需使用 Mac 上预先编译的 bin 文件。
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/crawler ./cmd/crawler

# 阶段二：准备运行环境，只从编译阶段复制可执行文件。
FROM alpine:3.24.4

# 爬虫通过 HTTPS 请求掘金，需要系统 CA 证书。
RUN apk add --no-cache ca-certificates \
    && mkdir -p /data \
    && chown 10001:10001 /data

COPY --from=builder /out/crawler /usr/local/bin/crawler
WORKDIR /data
USER 10001:10001

# 现有程序通过 signal.NotifyContext 监听 os.Interrupt（SIGINT）。
STOPSIGNAL SIGINT
ENTRYPOINT ["/usr/local/bin/crawler"]
# 无参数时显示帮助；docker run 镜像名 后的参数会替换这里的默认参数。
CMD ["-h"]
