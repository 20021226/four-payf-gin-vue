#!/usr/bin/env bash
set -euo pipefail

# 解析命令行参数
REBUILD=false
HELP=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --rebuild|-r)
      REBUILD=true
      shift
      ;;
    --help|-h)
      HELP=true
      shift
      ;;
    *)
      echo "未知参数: $1"
      echo "使用 --help 查看帮助信息"
      exit 1
      ;;
  esac
done

# 显示帮助信息
if [[ "$HELP" == true ]]; then
  echo "用法: $0 [选项]"
  echo ""
  echo "选项:"
  echo "  --rebuild, -r    强制重新构建所有镜像（源码更改后使用）"
  echo "  --help, -h       显示此帮助信息"
  echo ""
  echo "示例:"
  echo "  $0               # 正常启动服务"
  echo "  $0 --rebuild     # 重新构建镜像并启动服务"
  exit 0
fi

# 切换到脚本所在目录后，使用相对路径
cd "$(dirname "$0")"
COMPOSE_FILE="../docker-compose.yaml"
ENV_FILE="../.env"

if [[ ! -f "$COMPOSE_FILE" ]]; then
  echo "未找到 Compose 文件: $COMPOSE_FILE" >&2
  exit 1
fi

# 加载环境变量
if [[ ! -f "$ENV_FILE" ]]; then
  echo "未找到 .env 文件: $ENV_FILE" >&2
  exit 1
fi

echo "正在加载环境变量..."
source "$ENV_FILE"

# 设置默认值
SERVER_IP="${SERVER_IP:-127.0.0.1}"
DOMAIN="${DOMAIN:-example.com}"

echo "使用 Compose 文件: $COMPOSE_FILE"
echo "服务器IP: $SERVER_IP"

# 检查服务是否已经启动
echo "检查服务状态..."

# 兼容 docker-compose 与 docker compose 两种调用方式
if command -v docker-compose >/dev/null 2>&1; then
  COMPOSE_CMD="docker-compose"
else
  COMPOSE_CMD="docker compose"
fi

# 检查是否有运行中的服务
RUNNING_SERVICES=$($COMPOSE_CMD -f "$COMPOSE_FILE" ps --services --filter "status=running" 2>/dev/null || true)

if [[ -n "$RUNNING_SERVICES" ]]; then
  echo "发现运行中的服务，正在关闭..."
  $COMPOSE_CMD -f "$COMPOSE_FILE" down
  echo "服务已关闭"
fi

if [[ "$REBUILD" == true ]]; then
  echo "正在清理旧镜像和容器..."
  # 清理停止的容器
  docker container prune -f 2>/dev/null || true
  # 清理未使用的镜像
  docker image prune -f 2>/dev/null || true
  
  echo "正在重新构建镜像并启动服务..."
  $COMPOSE_CMD -f "$COMPOSE_FILE" up -d --build --force-recreate
  echo "镜像重新构建完成，服务启动完成"
else
  echo "正在启动服务..."
  $COMPOSE_CMD -f "$COMPOSE_FILE" up -d
  echo "服务启动完成"
fi

echo ""
echo "=========================================="
echo "           服务启动完成"
echo "=========================================="
echo "管理后台:        https://$SERVER_IP"
echo "默认账号:        admin"
echo "默认密码:        6830125352"
echo "API调用地址:     https://$SERVER_IP:8080/api/"
echo "=========================================="
