# 声明命令，就不会检查同名文件了，例如 make test 不会当成
.PHONY: help setup fmt lint test run build build-linux-arm64 build-linux-amd64

# 直接 make 不指定命令会执行第一条
# 命令冒号后面写其他命令，表示先执行其他命令

help: ## 显示所有可用命令
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# help:
# 	@echo 'make setup   安装 git hooks'
# 	@echo 'make fmt     格式化代码'
# 	@echo 'make lint    静态检查'
# 	@echo 'make test    运行测试'
# 	@echo 'make run     运行爬虫'
# 	@echo 'make build   编译到 bin/crawler 目录'

setup: ## 初始化 git hook
	lefthook install

fmt: ## 格式化代码
	golangci-lint fmt

lint: ## 静态检查
	golangci-lint run

# 运行测试 -race 开启数据竞争检测 -cover 输出测试覆盖率
# ./... 当前目录及所有子目录下的包，通配写法，相当于 **/**
test: ## 运行测试
	go test -race -cover ./...

# 编译并立刻运行，不留下可执行文件，-max 2 传入参数
run: ## 运行爬虫
	go run ./cmd/crawler -max 2

# 打包 -o bin/crawler 输出的可执行文件放在指定目录
# ./cmd/crawler 要编译的是这个目录下的 main 包
build: ## 编译构建到 bin/crawler
	go build -o bin/crawler ./cmd/crawler

build-linux-arm64: ## go tool dist list 查看支持的打包配置
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/crawler-linux-arm64 ./cmd/crawler

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/crawler-linux-amd64 ./cmd/crawler

# 默认串行执行，make -j3 build-all 最多 3 个并行执行
build-all: build-linux-arm64 build-linux-amd64
