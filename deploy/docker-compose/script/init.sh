#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

ROOT_DIR=".."  # 相对 deploy/docker-compose/script 到 compose 根
TEMPLATE_DIR="$ROOT_DIR/web/conf.d.template/sites"
OUTPUT_DIR="$ROOT_DIR/web/conf.d"
ENV_FILE="$ROOT_DIR/.env"
SSL_SRC_DIR="$ROOT_DIR/ssl"
SSL_TARGET_DIR="$ROOT_DIR/web/ssl"

chmod +x ../server/server
chmod +x ../web/entrypoint.sh

# 创建必要的数据目录
echo "正在创建必要的数据目录..."
DATA_DIRS=("/data/mysql" "/data/redis")

for dir in "${DATA_DIRS[@]}"; do
  if [[ ! -d "$dir" ]]; then
    echo "  创建目录: $dir"
    sudo mkdir -p "$dir" || {
      echo "创建目录失败: $dir" >&2
      echo "请手动执行: sudo mkdir -p $dir" >&2
      exit 1
    }
    # 设置适当的权限，让当前用户可以访问
    sudo chown -R $(whoami):$(whoami) "$dir" || {
      echo "设置目录权限失败: $dir，但目录已创建" >&2
    }
    echo "  目录创建成功: $dir"
  else
    echo "  目录已存在: $dir"
  fi
done

if [[ ! -f "$ENV_FILE" ]]; then
  echo "未找到 .env 文件: $ENV_FILE" >&2
  exit 1
fi

# 从 .env 读取 DOMAIN_NAME
source "$ENV_FILE"
DOMAIN_NAME="${DOMAIN_NAME:-${DOMAIN:-$SERVER_IP}}"
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

# 自动生成基于IP地址的SSL证书
echo "正在生成基于IP地址的SSL证书..."

# 获取服务器IP地址
SERVER_IP=$(hostname -I | awk '{print $1}')
if [[ -z "$SERVER_IP" ]]; then
  echo "警告: 无法获取服务器IP地址，使用默认IP 127.0.0.1"
  SERVER_IP="127.0.0.1"
fi

echo "检测到服务器IP: $SERVER_IP"

# 创建SSL证书配置文件
SSL_CONFIG_FILE="/tmp/ssl_config.conf"
cat > "$SSL_CONFIG_FILE" << EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C=CN
ST=Beijing
L=Beijing
O=Four-Pay
OU=IT Department
CN=$SERVER_IP

[v3_req]
basicConstraints = CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names

[alt_names]
IP.1 = $SERVER_IP
IP.2 = 127.0.0.1
DNS.1 = localhost
EOF

# 生成私钥和证书
echo "正在生成SSL私钥..."
openssl genrsa -out "$SSL_TARGET_DIR/server.key" 2048

echo "正在生成SSL证书..."
openssl req -new -x509 -key "$SSL_TARGET_DIR/server.key" \
  -out "$SSL_TARGET_DIR/server.pem" \
  -days 365 \
  -config "$SSL_CONFIG_FILE" \
  -extensions v3_req

# 设置适当的权限
chmod 600 "$SSL_TARGET_DIR/server.key"
chmod 644 "$SSL_TARGET_DIR/server.pem"
# 清理临时文件
rm -f "$SSL_CONFIG_FILE"

echo "SSL证书生成完成:"
echo "  证书文件: $SSL_TARGET_DIR/server.pem"
echo "  私钥文件: $SSL_TARGET_DIR/server.key"
echo "  证书有效期: 365天"
echo "  支持IP地址: $SERVER_IP, 127.0.0.1"


echo "所有配置文件已成功渲染到 $OUTPUT_DIR"
echo "输出目录: $OUTPUT_DIR"
echo "域名: $DOMAIN_NAME"


# 替换指定JS文件中的VITE_BASE_PATH配置
replace_specific_js_files() {
  local file1="087AC4D233B64EB0index.DaNEjAwX.js"
  local file2="087AC4D233B64EB0index-legacy.Cvng3aMO.js"
  local domain_name="$DOMAIN_NAME" 
  
  local assets_dir="$ROOT_DIR/web/html/assets"
  local new_base_path="https://$domain_name"
  
  echo "正在替换指定JS文件中的VITE_BASE_PATH配置..."
  echo "  目标地址: $domain_name"
  echo "  新的BASE_PATH: $new_base_path"
  
  # 处理第一个文件
  if [[ -f "$assets_dir/$file1" ]]; then
    echo "  处理文件: $file1"
    # 备份原文件
    cp "$assets_dir/$file1" "$assets_dir/$file1.backup.$(date +%Y%m%d_%H%M%S)"
    
    # 替换VITE_BASE_PATH配置
    sed -i "s|VITE_BASE_PATH:\"[^\"]*\"|VITE_BASE_PATH:\"$new_base_path\"|g" "$assets_dir/$file1"
    # sed -i "s|VITE_BASE_API:\"[^\"]*\"|VITE_BASE_API:\"/api\"|g" "$assets_dir/$file1"
    # sed -i "s|https://demo\.gin-vue-admin\.com|$new_base_path|g" "$assets_dir/$file1"
    
    echo "    ✓ $file1 配置已更新"
  else
    echo "    ⚠️  文件不存在: $file1"
  fi
  
  # 处理第二个文件
  if [[ -f "$assets_dir/$file2" ]]; then
    echo "  处理文件: $file2"
    # 备份原文件
    cp "$assets_dir/$file2" "$assets_dir/$file2.backup.$(date +%Y%m%d_%H%M%S)"
    
    # 替换VITE_BASE_PATH配置
    sed -i "s|VITE_BASE_PATH:\"[^\"]*\"|VITE_BASE_PATH:\"$new_base_path\"|g" "$assets_dir/$file2"
    sed -i "s|VITE_BASE_API:\"[^\"]*\"|VITE_BASE_API:\"/api\"|g" "$assets_dir/$file2"
    sed -i "s|https://demo\.gin-vue-admin\.com|$new_base_path|g" "$assets_dir/$file2"
    
    echo "    ✓ $file2 配置已更新"
  else
    echo "    ⚠️  文件不存在: $file2"
  fi
  
  echo "  指定文件的VITE配置替换完成"
}

# 调用JS文件配置替换方法
replace_specific_js_files

echo "Nginx 配置初始化完成"