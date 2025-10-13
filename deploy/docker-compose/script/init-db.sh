#!/bin/bash

# 配置参数
MYSQL_CONTAINER_NAME="g-mysql"
MYSQL_ROOT_USER="root"
MYSQL_ROOT_PASSWORD="123456"
MYSQL_PORT=13306
SQL_FILE_PATH="../mysql/init.sql"   # SQL 文件路径
TARGET_DB="four_pay"           # 目标数据库名

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

# 等待 MySQL 启动
echo "等待 MySQL 启动..."
sleep 10

# 用户确认删除数据库
read -p "是否删除数据库 $TARGET_DB 并重新创建？(yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo "操作已取消"
    exit 0
fi

# 删除数据库
docker exec -i $MYSQL_CONTAINER_NAME mysql -u$MYSQL_ROOT_USER -p$MYSQL_ROOT_PASSWORD -e "DROP DATABASE IF EXISTS $TARGET_DB;"
echo "数据库 $TARGET_DB 已删除"

# 创建数据库
docker exec -i $MYSQL_CONTAINER_NAME mysql -u$MYSQL_ROOT_USER -p$MYSQL_ROOT_PASSWORD -e "CREATE DATABASE $TARGET_DB;"
echo "数据库 $TARGET_DB 已创建"

# 执行 SQL 文件
docker exec -i $MYSQL_CONTAINER_NAME mysql -u$MYSQL_ROOT_USER -p$MYSQL_ROOT_PASSWORD $TARGET_DB < $SQL_FILE_PATH
echo "SQL 文件执行完成！"