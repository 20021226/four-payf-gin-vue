#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

ROOT_DIR=".."  # 相对 deploy/docker-compose/script 到 compose 根
TEMPLATE_DIR="$ROOT_DIR/nginx/conf.d.template/sites"
OUTPUT_DIR="$ROOT_DIR/nginx/conf.d/sites"
ENV_FILE="$ROOT_DIR/.env"
SSL_SRC_DIR="$ROOT_DIR/ssl"
SSL_TARGET_DIR="$ROOT_DIR/nginx/ssl"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "未找到 .env 文件: $ENV_FILE" >&2
  exit 1
fi

# 从 .env 读取 DOMAIN_NAME
source "$ENV_FILE"
DOMAIN_NAME="${DOMAIN_NAME:-${DOMAIN:-example.com}}"
echo "使用域名: $DOMAIN_NAME"

mkdir -p "$OUTPUT_DIR"
mkdir -p "$SSL_TARGET_DIR"

# 批量渲染 sites 目录中的所有模板
if [[ -d "$TEMPLATE_DIR" ]]; then
  for template_file in "$TEMPLATE_DIR"/*.conf; do
    if [[ -f "$template_file" ]]; then
      filename="$(basename "$template_file")"
      echo "  处理文件: $filename"
      DOMAIN_NAME="$DOMAIN_NAME" envsubst '${DOMAIN_NAME}' < "$template_file" > "$OUTPUT_DIR/$filename"
    fi
  done
else
  echo "模板目录不存在: $TEMPLATE_DIR" >&2
fi

# 复制证书到 nginx/ssl 并重命名为 server.pem 与 server.key
if [[ -d "$SSL_SRC_DIR" ]]; then
  cert_candidate=$(ls "$SSL_SRC_DIR"/*.pem "$SSL_SRC_DIR"/*.crt 2>/dev/null | head -n1 || true)
  if [[ -n "${cert_candidate:-}" ]]; then
    cp -f "$cert_candidate" "$SSL_TARGET_DIR/server.pem"
  fi
  key_candidate=$(ls "$SSL_SRC_DIR"/*.key 2>/dev/null | head -n1 || true)
  if [[ -n "${key_candidate:-}" ]]; then
    cp -f "$key_candidate" "$SSL_TARGET_DIR/server.key"
  fi
  echo "已从 $SSL_SRC_DIR 复制证书到 $SSL_TARGET_DIR 并重命名为 server.pem/server.key"
else
  echo "未找到证书源目录: $SSL_SRC_DIR"
fi


echo "已生成并复制到: $OUTPUT_DIR"
echo "Nginx 配置初始化完成"