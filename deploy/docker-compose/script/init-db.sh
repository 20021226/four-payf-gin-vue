#!/usr/bin/env bash
set -euo pipefail

# 配置参数
MYSQL_CONTAINER_NAME="gva-mysql"
MYSQL_ROOT_USER="root"
MYSQL_ROOT_PASSWORD="123456"
MYSQL_PORT=13306
SQL_FILE_PATH="./mysql/init.sql"   # SQL 文件路径
TARGET_DB="four_pay"           # 目标数据库名

# Redis 配置参数
REDIS_CONTAINER_NAME="gva-redis"
REDIS_PORT=16379

MYSQL_DATA_DIR="/data/mysql"
REDIS_DATA_DIR="/data/redis"

# 检查 MySQL 数据目录是否存在
if [ ! -d "$MYSQL_DATA_DIR" ]; then
    echo "MySQL 数据目录不存在，正在创建: $MYSQL_DATA_DIR"
    mkdir -p "$MYSQL_DATA_DIR"
else
    echo "MySQL 数据目录已存在: $MYSQL_DATA_DIR"
fi

# 检查 Redis 数据目录是否存在
if [ ! -d "$REDIS_DATA_DIR" ]; then
    echo "Redis 数据目录不存在，正在创建: $REDIS_DATA_DIR"
    mkdir -p "$REDIS_DATA_DIR"
else
    echo "Redis 数据目录已存在: $REDIS_DATA_DIR"
fi

# 检查 SQL 目录是否存在，存在则删除重新创建
if [ -d "$SQL_DIR_PATH" ]; then
    echo "目录已存在，删除并重新创建: $SQL_DIR_PATH"
    rm -rf "$SQL_DIR_PATH"
fi


# 检查 SQL 文件是否存在
if [ ! -f "$SQL_FILE_PATH" ]; then
    echo "SQL 文件不存在: $SQL_FILE_PATH"
    exit 1
fi

# 启动 MySQL 服务（如果未运行则创建）
if [ "$(docker compose ps -q mysql)" ]; then
    echo "MySQL 服务已存在，启动中..."
    docker compose start mysql
else
    echo "创建并启动 MySQL 服务..."
    docker compose up -d mysql
fi

# 启动 Redis 服务（如果未运行则创建）
if [ "$(docker compose ps -q redis)" ]; then
    echo "Redis 服务已存在，启动中..."
    docker compose start redis
else
    echo "创建并启动 Redis 服务..."
    docker compose up -d redis
fi

# 等待服务启动
echo "等待 MySQL 和 Redis 启动..."
sleep 10

# 用户确认删除数据库和Redis数据
echo "=========================================="
echo "警告：此操作将会："
echo "1. 删除 MySQL 数据库: $TARGET_DB"
echo "2. 清空 Redis 所有数据"
echo "3. 重新初始化数据库"
echo "=========================================="
read -p "是否继续执行？(yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo "操作已取消"
    exit 0
fi

# 删除数据库
echo "正在删除 MySQL 数据库..."
docker exec -i $MYSQL_CONTAINER_NAME mysql -u$MYSQL_ROOT_USER -p$MYSQL_ROOT_PASSWORD -e "DROP DATABASE IF EXISTS $TARGET_DB;"
echo "✓ 数据库 $TARGET_DB 已删除"

# 清空 Redis 数据
echo "正在清空 Redis 数据..."
docker exec -i $REDIS_CONTAINER_NAME redis-cli FLUSHALL
echo "✓ Redis 数据已清空"

# 创建数据库
echo "正在创建 MySQL 数据库..."
docker exec -i $MYSQL_CONTAINER_NAME mysql -u$MYSQL_ROOT_USER -p$MYSQL_ROOT_PASSWORD -e "CREATE DATABASE $TARGET_DB;"
echo "✓ 数据库 $TARGET_DB 已创建"

# 执行 SQL 文件
echo "正在执行 SQL 初始化文件..."
docker exec -i $MYSQL_CONTAINER_NAME mysql -u$MYSQL_ROOT_USER -p$MYSQL_ROOT_PASSWORD $TARGET_DB < $SQL_FILE_PATH
echo "✓ SQL 文件执行完成！"

echo ""
echo "=========================================="
echo "           数据初始化完成"
echo "=========================================="
echo "MySQL 数据库: $TARGET_DB (已重新创建)"
echo "Redis 缓存:   已清空"
echo "=========================================="