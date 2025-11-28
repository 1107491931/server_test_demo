#!/bin/bash

echo "================================"
echo "Multi-Environment Verification"
echo "================================"
echo ""

# 1. 验证 hosts 配置
echo "1. Checking hosts configuration..."
echo "--------------------------------"
for domain in api.dev.myapp.com api.staging.myapp.com api.pre.myapp.com api.prod.myapp.com; do
    if grep -q "$domain" /etc/hosts 2>/dev/null; then
        echo "✅ $domain found in /etc/hosts"
    else
        echo "❌ $domain NOT found in /etc/hosts"
    fi
done

# 2. 验证 DNS 解析
echo ""
echo "2. Testing DNS resolution..."
echo "--------------------------------"
for domain in api.dev.myapp.com api.staging.myapp.com api.pre.myapp.com api.prod.myapp.com; do
    if ping -c 1 -t 1 "$domain" > /dev/null 2>&1; then
        echo "✅ $domain → 127.0.0.1"
    else
        echo "❌ $domain → Resolution failed"
    fi
done

# 3. 验证 Docker 容器
echo ""
echo "3. Checking Docker containers..."
echo "--------------------------------"
docker ps --filter "name=test_demo" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 4. 验证端口访问
echo ""
echo "4. Testing port access (direct)..."
echo "--------------------------------"
PORTS=(8081 8082 8083 8084)
ENVS=(dev staging pre prod)

for i in "${!PORTS[@]}"; do
    port=${PORTS[$i]}
    env=${ENVS[$i]}
    
    if curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:$port/users | grep -q "200\|404"; then
        echo "✅ Port $port ($env) is responding"
    else
        echo "❌ Port $port ($env) is not responding"
    fi
done

# 5. 验证域名访问（通过 Nginx）
echo ""
echo "5. Testing domain access (via Nginx)..."
echo "--------------------------------"
for env in dev staging pre prod; do
    if curl -s -o /dev/null -w "%{http_code}" http://api.$env.myapp.com/users | grep -q "200\|404"; then
        echo "✅ api.$env.myapp.com is working"
    else
        echo "⚠️  api.$env.myapp.com is not working"
    fi
done

# 6. 验证 Nginx 状态
echo ""
echo "6. Checking Nginx status..."
echo "--------------------------------"
if ps aux | grep -v grep | grep -q nginx; then
    echo "✅ Nginx is running"
    
    # 检查配置文件
    if [ -f "/usr/local/etc/nginx/servers/testdemo.conf" ]; then
        echo "✅ Nginx config found at /usr/local/etc/nginx/servers/testdemo.conf"
    elif [ -f "/etc/nginx/conf.d/testdemo.conf" ]; then
        echo "✅ Nginx config found at /etc/nginx/conf.d/testdemo.conf"
    else
        echo "⚠️  Nginx config not found"
    fi
else
    echo "⚠️  Nginx is not running"
fi

# 7. 验证数据库目录
echo ""
echo "7. Checking database directories..."
echo "--------------------------------"
for env in dev staging pre prod; do
    if [ -d "dbs/$env" ]; then
        echo "✅ dbs/$env directory exists"
    else
        echo "❌ dbs/$env directory not found"
    fi
done

# 总结
echo ""
echo "================================"
echo "Summary"
echo "================================"
echo ""
echo "📋 Access URLs:"
echo "   Dev:     http://api.dev.myapp.com"
echo "   Staging: http://api.staging.myapp.com"
echo "   Pre:     http://api.pre.myapp.com"
echo "   Prod:    http://api.prod.myapp.com"
echo ""
echo "📚 Swagger:"
echo "   http://api.staging.myapp.com/swagger/index.html"
echo ""
echo "🔍 Troubleshooting:"
echo "   - If domain access fails, check Nginx: sudo nginx -t"
echo "   - If port access fails, check containers: docker ps"
echo "   - View logs: docker logs test_demo_staging"
echo "   - Nginx logs: tail -f /usr/local/var/log/nginx/testdemo_staging_error.log"
