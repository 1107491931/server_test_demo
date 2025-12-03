#!/bin/bash

# 速率限制测试脚本
# 用于测试 user-service 的速率限制功能

echo "========================================="
echo "速率限制测试脚本"
echo "========================================="
echo ""

# 配置
SERVICE_URL="http://localhost:8081"
TOTAL_REQUESTS=20

echo "测试配置:"
echo "  服务地址: $SERVICE_URL"
echo "  请求总数: $TOTAL_REQUESTS"
echo ""

# 测试 1: 测试登录接口的严格限流 (每秒2个请求，突发5个)
echo "测试 1: 登录接口限流测试"
echo "----------------------------------------"
echo "预期: 前5个请求成功，后续请求会被限流"
echo ""

SUCCESS_COUNT=0
RATE_LIMITED_COUNT=0

for i in $(seq 1 $TOTAL_REQUESTS); do
  RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    "$SERVICE_URL/api/v1/users/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"test$i\",\"password\":\"123456\"}")
  
  HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
  
  if [ "$HTTP_CODE" = "429" ]; then
    echo "请求 $i: ❌ 被限流 (429)"
    ((RATE_LIMITED_COUNT++))
  else
    echo "请求 $i: ✅ 成功 ($HTTP_CODE)"
    ((SUCCESS_COUNT++))
  fi
  
  # 每个请求之间间隔很短，以触发限流
  sleep 0.05
done

echo ""
echo "测试结果:"
echo "  成功请求: $SUCCESS_COUNT"
echo "  被限流请求: $RATE_LIMITED_COUNT"
echo ""

# 等待一段时间让令牌桶恢复
echo "等待 3 秒让令牌桶恢复..."
sleep 3
echo ""

# 测试 2: 测试查询接口的一般限流 (每秒10个请求，突发20个)
echo "测试 2: 查询接口限流测试"
echo "----------------------------------------"
echo "预期: 前20个请求成功，后续请求会被限流"
echo ""

SUCCESS_COUNT=0
RATE_LIMITED_COUNT=0

for i in $(seq 1 $TOTAL_REQUESTS); do
  RESPONSE=$(curl -s -w "\n%{http_code}" -X GET \
    "$SERVICE_URL/api/v1/users")
  
  HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
  
  if [ "$HTTP_CODE" = "429" ]; then
    echo "请求 $i: ❌ 被限流 (429)"
    ((RATE_LIMITED_COUNT++))
  else
    echo "请求 $i: ✅ 成功 ($HTTP_CODE)"
    ((SUCCESS_COUNT++))
  fi
  
  sleep 0.05
done

echo ""
echo "测试结果:"
echo "  成功请求: $SUCCESS_COUNT"
echo "  被限流请求: $RATE_LIMITED_COUNT"
echo ""

echo "========================================="
echo "测试完成"
echo "========================================="
