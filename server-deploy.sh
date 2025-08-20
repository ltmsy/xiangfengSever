#!/bin/bash

# 服务器端部署脚本
# 在服务器上执行

echo "=== 唐僧叨叨生产环境部署脚本 ==="

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker服务"
    exit 1
fi

# 检查Docker Compose是否可用
if ! docker-compose --version > /dev/null 2>&1; then
    echo "❌ Docker Compose不可用，请先安装"
    exit 1
fi

echo "✅ Docker环境检查通过"

# 创建必要的目录
mkdir -p logs/tsdd logs/wk

# 设置目录权限
chmod 755 logs/tsdd logs/wk

echo "✅ 目录创建完成"

# 构建Docker镜像
echo "🔨 开始构建Docker镜像..."
make build

if [ $? -ne 0 ]; then
    echo "❌ 构建失败，请检查错误信息"
    exit 1
fi

echo "✅ Docker镜像构建成功"

# 停止现有服务
echo "🛑 停止现有服务..."
docker-compose -f docker-compose-prod.yaml down 2>/dev/null || true

# 启动服务
echo "🚀 启动生产服务..."
docker-compose -f docker-compose-prod.yaml up -d

if [ $? -eq 0 ]; then
    echo "✅ 服务启动成功！"
    
    # 等待服务启动
    echo "⏳ 等待服务启动..."
    sleep 10
    
    # 检查服务状态
    echo "📊 服务状态检查:"
    docker-compose -f docker-compose-prod.yaml ps
    
    echo ""
    echo "🎉 部署完成！"
    echo ""
    echo "📱 服务访问地址:"
    echo "  - 唐僧叨叨API: http://112.121.164.130:8090"
    echo "  - 悟空IM API: http://112.121.164.130:5001"
    echo "  - Web界面: http://112.121.164.130:82"
    echo "  - MinIO控制台: http://112.121.164.130:9001"
    echo ""
    echo "🔧 管理命令:"
    echo "  - 查看日志: docker-compose -f docker-compose-prod.yaml logs -f"
    echo "  - 停止服务: docker-compose -f docker-compose-prod.yaml down"
    echo "  - 重启服务: docker-compose -f docker-compose-prod.yaml restart"
    
else
    echo "❌ 服务启动失败，请检查日志"
    docker-compose -f docker-compose-prod.yaml logs
    exit 1
fi 