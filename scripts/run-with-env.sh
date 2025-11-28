#!/bin/bash

# 加载环境变量文件并运行应用
# 使用方法: ./scripts/run-with-env.sh staging

ENV_NAME=$1

if [ -z "$ENV_NAME" ]; then
    echo "用法: $0 {dev|staging|pre|prod}"
    exit 1
fi

ENV_FILE="config/.env.$ENV_NAME"

if [ ! -f "$ENV_FILE" ]; then
    echo "错误: 配置文件 $ENV_FILE 不存在"
    exit 1
fi

echo "========================================="
echo "加载环境配置: $ENV_FILE"
echo "========================================="
cat "$ENV_FILE"
echo "========================================="
echo ""

# 导出环境变量
export $(cat "$ENV_FILE" | grep -v '^#' | xargs)

# 运行应用
go run main.go
