# 声明命令，就不会检查同名文件了，例如 make test 不会当成
.PHONY: fmt lint test build run

# 直接 make 不指定命令会执行第一条
# 命令冒号后面写其他命令，表示先执行其他命令

fmt:
	golangci-lint fmt

lint:
	golangci-lint run

# 运行测试 -race 开启数据竞争检测 -cover 输出测试覆盖率
# ./... 当前目录及所有子目录下的包，通配写法，相当于 **/**
test:
	go test -race -cover ./...

# 打包 -o bin/crawler 输出的可执行文件放在指定目录
# ./cmd/crawler 要编译的是这个目录下的 main 包
build:
	go build -o bin/crawler ./cmd/crawler

# 编译并立刻运行，不留下可执行文件，-max 2 传入参数
run:
	go run ./cmd/crawler -max 2
