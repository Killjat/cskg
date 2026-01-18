FROM alpine:3.18

# 设置工作目录
WORKDIR /app

# 复制二进制文件
COPY nodemanage /app/

# 复制配置文件
COPY config.yaml /app/

# 安装必要的依赖
RUN apk update && apk add --no-cache bash

# 设置默认命令
CMD ["./nodemanage", "client", "-n", "node-1"]
