## 前提条件
- 系统使用 ubuntu 24.04
- 安装 [Docker]()
- 准备域名和域名对应的ssl证书
- 配置域名解析到当前要部署的服务器ip

## 安装 docker
### 国内的服务器
```shell
curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun
```
### 国外的服务器
```shell
curl -fsSL https://get.docker.com -o get-docker.sh | base64 -d > get-docker.sh
chmod +x get-docker.sh
./get-docker.sh
```

## 证书配置
- 证书文件需要放在`docker-compose/ssl/`目录下

## env 文件配置
- 配置文件示例:
```shell
DOMAIN=example.com # 服务使用的域名
```

## 初始化数据库 (运行一次即可)
```shell
./script/initdb.sh
```

## 1.初始化
```shell
./script/init.sh
```
## 2. 启动
```shell
./script/start.sh
```
## 注意点
- 每次运行`./script/init.sh`初始化脚本时，会删除所有容器和镜像，包括数据库容器。
- 每次运行`./script/start.sh`启动脚本时，会启动所有容器，包括数据库容器。
- 每次运行`./script/initdb.sh`初始化数据库脚本时，会初始化数据库容器, 原来的数据将丢失。


