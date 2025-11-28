#!/bin/bash

echo "================================"
echo "Multi-Environment Setup Script"
echo "================================"
echo ""

# 检查是否在项目根目录
if [ ! -f "main.go" ]; then
    echo "❌ Error: Please run this script from the project root directory"
    exit 1
fi

# 1. 配置 hosts
echo "Step 1: Configuring hosts file..."
if ! grep -q "api.dev.myapp.com" /etc/hosts; then
    echo "Adding entries to /etc/hosts (requires sudo)..."
    sudo bash -c 'cat >> /etc/hosts << EOF

# Test Demo Project - Multi Environment
127.0.0.1  api.dev.myapp.com
127.0.0.1  api.staging.myapp.com
127.0.0.1  api.pre.myapp.com
127.0.0.1  api.prod.myapp.com
EOF'
    echo "✅ Hosts configured"
else
    echo "✅ Hosts already configured"
fi

# 2. 清除 DNS 缓存
echo ""
echo "Step 2: Flushing DNS cache..."
sudo dscacheutil -flushcache 2>/dev/null
sudo killall -HUP mDNSResponder 2>/dev/null
echo "✅ DNS cache cleared"

# 3. 验证 hosts 配置
echo ""
echo "Step 3: Verifying hosts configuration..."
for domain in api.dev.myapp.com api.staging.myapp.com api.pre.myapp.com api.prod.myapp.com; do
    if ping -c 1 -t 1 "$domain" > /dev/null 2>&1; then
        echo "✅ $domain → 127.0.0.1"
    else
        echo "❌ $domain → Resolution failed"
    fi
done

# 4. 检查 Nginx
echo ""
echo "Step 4: Checking Nginx..."
if command -v nginx > /dev/null 2>&1; then
    echo "✅ Nginx is installed ($(nginx -v 2>&1))"
    
    # 复制配置文件
    if [ -f "nginx/testdemo.conf" ]; then
        echo "Copying Nginx configuration..."
        if [ -d "/usr/local/etc/nginx/servers" ]; then
            sudo cp nginx/testdemo.conf /usr/local/etc/nginx/servers/testdemo.conf
            echo "✅ Nginx config copied to /usr/local/etc/nginx/servers/"
        elif [ -d "/etc/nginx/conf.d" ]; then
            sudo cp nginx/testdemo.conf /etc/nginx/conf.d/testdemo.conf
            echo "✅ Nginx config copied to /etc/nginx/conf.d/"
        fi
        
        # 测试配置
        if sudo nginx -t 2>&1 | grep -q "successful"; then
            echo "✅ Nginx configuration is valid"
            sudo nginx -s reload 2>/dev/null || sudo nginx
            echo "✅ Nginx reloaded"
        else
            echo "⚠️  Nginx configuration test failed, please check manually"
        fi
    fi
else
    echo "⚠️  Nginx is not installed"
    echo "   Install with: brew install nginx (macOS)"
    echo "   Or: sudo apt install nginx (Ubuntu/Debian)"
fi

# 5. 创建目录
echo ""
echo "Step 5: Creating directories..."
mkdir -p config dbs/dev dbs/staging dbs/pre dbs/prod
echo "✅ Directories created"

# 6. 检查 Docker
echo ""
echo "Step 6: Checking Docker..."
if command -v docker > /dev/null 2>&1; then
    echo "✅ Docker is installed ($(docker --version))"
else
    echo "❌ Docker is not installed"
    echo "   Please install Docker Desktop from: https://www.docker.com/products/docker-desktop"
    exit 1
fi

# 7. 检查镜像
echo ""
echo "Step 7: Checking Docker image..."
if docker images | grep -q "test_demo_1.0.0"; then
    echo "✅ Docker image 'test_demo_1.0.0' found"
else
    echo "⚠️  Docker image 'test_demo_1.0.0' not found"
    echo "   Building image..."
    docker build -t test_demo_1.0.0 .
    if [ $? -eq 0 ]; then
        echo "✅ Image built successfully"
    else
        echo "❌ Image build failed"
        exit 1
    fi
fi

# 8. 启动容器
echo ""
echo "Step 8: Starting Docker containers..."
if command -v docker-compose > /dev/null 2>&1; then
    echo "Using docker-compose..."
    docker-compose up -d
    echo "✅ Containers started with docker-compose"
else
    echo "docker-compose not found, using docker run..."
    
    # 停止并删除旧容器
    for env in dev staging pre prod; do
        docker stop test_demo_$env 2>/dev/null
        docker rm test_demo_$env 2>/dev/null
    done
    
    # 启动新容器
    docker run -d --name test_demo_dev -p 8081:8081 --env-file config/.env.dev -v $(pwd)/dbs/dev:/app/dbs test_demo_1.0.0
    docker run -d --name test_demo_staging -p 8082:8081 --env-file config/.env.staging -v $(pwd)/dbs/staging:/app/dbs test_demo_1.0.0
    docker run -d --name test_demo_pre -p 8083:8081 --env-file config/.env.pre -v $(pwd)/dbs/pre:/app/dbs test_demo_1.0.0
    docker run -d --name test_demo_prod -p 8084:8081 --env-file config/.env.prod -v $(pwd)/dbs/prod:/app/dbs test_demo_1.0.0
    
    echo "✅ Containers started"
fi

# 9. 等待容器启动
echo ""
echo "Step 9: Waiting for containers to start..."
sleep 5

# 10. 验证
echo ""
echo "Step 10: Verifying setup..."
echo "Testing port access (direct)..."
for port in 8081 8082 8083 8084; do
    if curl -s http://127.0.0.1:$port/users > /dev/null 2>&1; then
        echo "✅ Port $port is responding"
    else
        echo "⚠️  Port $port is not responding"
    fi
done

echo ""
echo "Testing domain access (via Nginx)..."
for env in dev staging pre prod; do
    if curl -s http://api.$env.myapp.com/users > /dev/null 2>&1; then
        echo "✅ api.$env.myapp.com is working"
    else
        echo "⚠️  api.$env.myapp.com is not working (Nginx may not be configured)"
    fi
done

# 完成
echo ""
echo "================================"
echo "Setup Complete!"
echo "================================"
echo ""
echo "📋 Access your environments:"
echo "   Dev:     http://api.dev.myapp.com"
echo "   Staging: http://api.staging.myapp.com"
echo "   Pre:     http://api.pre.myapp.com"
echo "   Prod:    http://api.prod.myapp.com"
echo ""
echo "📚 Swagger Documentation:"
echo "   http://api.staging.myapp.com/swagger/index.html"
echo ""
echo "🐳 Docker Containers:"
docker ps --filter "name=test_demo" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""
echo "📖 For more details, see: a-docs/Multi_Environment_Setup.md"
