# 声明命令，就不会检查同名文件了，例如 make test 不会当成
.PHONY: help setup fmt lint test build run

# 直接 make 不指定命令会执行第一条
# 命令冒号后面写其他命令，表示先执行其他命令

help: ## 显示所有可用命令
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# help:
# 	@echo 'make setup   安装 git hooks'
# 	@echo 'make fmt     格式化代码'
# 	@echo 'make lint    静态检查'
# 	@echo 'make test    运行测试'
# 	@echo 'make build   编译到 bin/crawler 目录'
# 	@echo 'make run     运行爬虫'

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

# 打包 -o bin/crawler 输出的可执行文件放在指定目录
# ./cmd/crawler 要编译的是这个目录下的 main 包
build: ## 编译构建到 bin/crawler
	go build -o bin/crawler ./cmd/crawler

# 编译并立刻运行，不留下可执行文件，-max 2 传入参数
run: ## 运行爬虫
	go run ./cmd/crawler -max 2
