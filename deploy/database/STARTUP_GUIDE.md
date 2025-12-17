# Docker 数据库启动指南

## 推荐方式：使用 docker-compose

### 方法一：从项目根目录启动

```powershell
# 进入数据库部署目录
cd deploy/database

# 启动MySQL容器
docker-compose up -d

# 查看容器状态
docker ps

# 返回项目根目录
cd ../..
```

### 方法二：使用绝对路径

```powershell
# 直接在任意位置执行
docker-compose -f "d:\工程\jd-task-platform-go\deploy\database\docker-compose.yml" up -d
```

## 验证容器运行状态

```powershell
# 检查容器是否运行
docker ps --filter "name=jd-task-mysql"

# 查看容器日志
docker logs jd-task-mysql

# 进入容器
docker exec -it jd-task-mysql bash
```

## 停止和清理

```powershell
# 停止容器（保留数据）
cd deploy/database
docker-compose down

# 停止并删除数据卷（⚠️ 会删除所有数据）
docker-compose down -v
```

## 数据持久化说明

- 数据存储在 Docker 命名卷 `mysql-data` 中
- 即使删除容器，数据也不会丢失
- 只有执行 `docker-compose down -v` 才会删除数据

## 常见问题

### 端口被占用

如果3306端口被占用，可以修改 `docker-compose.yml` 中的端口映射：
```yaml
ports:
  - "3307:3306"  # 将主机端口改为3307
```

### 容器无法启动

检查Docker Desktop是否正在运行：
```powershell
docker info
```

### 数据迁移

如需迁移数据到新的 docker-compose 配置：
1. 备份当前数据
2. 停止旧容器
3. 使用新的 docker-compose.yml 启动
4. 数据会自动挂载到命名卷