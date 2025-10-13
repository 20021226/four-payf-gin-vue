#!/usr/bin/env bash
set -euo pipefail

# 切换到脚本所在目录后，使用相对路径
cd "$(dirname "$0")"
COMPOSE_FILE="../docker-compose.yaml"

if [[ ! -f "$COMPOSE_FILE" ]]; then
  echo "未找到 Compose 文件: $COMPOSE_FILE" >&2
  exit 1
fi

echo "使用 Compose 文件: $COMPOSE_FILE"

# 兼容 docker-compose 与 docker compose 两种调用方式
if command -v docker-compose >/dev/null 2>&1; then
  docker-compose -f "$COMPOSE_FILE" up -d
else
  docker compose -f "$COMPOSE_FILE" up -d
fi
