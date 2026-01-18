# 基础镜像
FROM golang:1.25.5-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制项目代码
COPY . .

# 构建扫描节点
RUN CGO_ENABLED=0 GOOS=linux go build -o scan_node cmd/scan_node/main.go

# 最终镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 复制构建好的扫描节点
COPY --from=builder /app/scan_node .

# 复制配置文件
COPY config/config.yaml ./config/

# 设置环境变量
ENV CONFIG_PATH=/app/config

# 运行扫描节点
CMD ["./scan_node", "-config", "/app/config"]