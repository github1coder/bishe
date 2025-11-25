FROM golang:1.24-bullseye AS builder

# RUN apk add --no-cache gcc libc-dev

# 为我们的镜像设置必要的环境变量
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

# 设置工作目录（统一用 /app，避免根目录混乱）
WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存加速后续构建
COPY go.mod go.sum ./
RUN go mod download  
# 提前下载依赖，代码变更时无需重复下载

# 复制项目所有代码
COPY . .



# 2. 复制项目所有源码（放在依赖之后，避免源码变更导致依赖重新下载）
COPY . .

# 暴露程序使用的端口（与你的程序端口一致，如 9000）
EXPOSE 9000

# 容器启动命令：通过 go run 直接执行源码（需指定程序入口文件，如 main.go）
# 注意：替换 ./main.go 为你的程序实际入口文件路径（如 ./cmd/server/main.go）
ENTRYPOINT ["go", "run", "./main.go", "conf/config.ini"]

# =====


# # 编译二进制文件，并赋予执行权限
# RUN go build -o chainqa_offchain_demo . && chmod +x chainqa_offchain_demo

# # 检查编译结果
# RUN ls -l chainqa_offchain_demo

# # 暴露应用可能用到的端口（根据你的实际端口修改，如 9000
# EXPOSE 9000

# # 启动命令（与原逻辑一致）
# ENTRYPOINT ["./chainqa_offchain_demo", "conf/config.ini"]