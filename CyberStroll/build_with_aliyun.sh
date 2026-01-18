#!/bin/bash

echo "使用阿里云镜像源构建Docker镜像..."

# 设置Docker镜像源环境变量
export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=plain

# 使用阿里云镜像源构建扫描节点镜像
docker build \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  --network host \
  -t cyberstroll-scan-node:latest .

echo "构建完成！镜像名称：cyberstroll-scan-node:latest"
echo "可以使用以下命令启动服务："
echo "docker-compose up -d"
