# 数据库部署配置

## 目录结构

- `docker-compose.yml`: Docker Compose 配置文件，用于启动 MySQL 数据库容器
- `database_schema.sql/`: 数据库结构定义目录
- `mysql-init/`: MySQL 初始化脚本目录

## 使用说明

1. 确保已安装 Docker 和 Docker Compose
2. 在当前目录下运行以下命令启动数据库：

```bash
docker-compose up -d
```

3. 数据库将运行在本地 3306 端口
   - 数据库名: jd
   - 用户名: jduser
   - 密码: jdpass123
   - Root 密码: 123456

## 数据持久化

数据库使用 Docker 命名卷进行数据持久化，即使容器被删除，数据也不会丢失。