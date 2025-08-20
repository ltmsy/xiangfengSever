# 🚀 唐僧叨叨生产环境部署

## 📋 部署步骤

### 1. 上传源码到服务器
```bash
# 在本地执行，上传整个项目到服务器
scp -r . root@112.121.164.130:/opt/tangsengdaodao/
```

### 2. 在服务器上部署
```bash
# 登录到服务器
ssh root@112.121.164.130

# 进入项目目录
cd /opt/tangsengdaodao

# 给部署脚本执行权限
chmod +x server-deploy.sh

# 执行部署
./server-deploy.sh
```

## 🔧 服务配置

### 端口配置
- **唐僧叨叨API**: 8090
- **悟空IM HTTP API**: 5001  
- **悟空IM TCP**: 5100
- **悟空IM WebSocket**: 5200
- **悟空IM 监控**: 5300
- **Web界面**: 82
- **GRPC Webhook**: 6979

### 外部服务
- **数据库**: MySQL (112.121.164.130:3306)
- **Redis**: 112.121.164.130:6379
- **MinIO**: http://112.121.164.130:9000

## 📱 访问地址
- 唐僧叨叨API: http://112.121.164.130:8090
- 悟空IM API: http://112.121.164.130:5001
- Web界面: http://112.121.164.130:82
- MinIO控制台: http://112.121.164.130:9001

## 🛠️ 管理命令
```bash
# 查看服务状态
docker-compose -f docker/tsdd/docker-compose.yaml ps

# 查看日志
docker-compose -f docker/tsdd/docker-compose.yaml logs -f

# 停止服务
docker-compose -f docker/tsdd/docker-compose.yaml down

# 重启服务
docker-compose -f docker/tsdd/docker-compose.yaml restart
```

## ⚠️ 注意事项
1. 确保服务器已安装Docker和Docker Compose
2. 确保防火墙开放相应端口
3. 确保外部服务(MySQL、Redis、MinIO)已启动并可访问
4. 生产环境建议配置HTTPS和SSL证书 