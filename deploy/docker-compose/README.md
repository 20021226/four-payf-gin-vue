## 前提条件
- 系统要求 [Ubuntu]() 或者 其他linux发行版
- 安装 [Docker]()
<!-- - 配置域名解析到当前要部署的服务器ip -->

## 安装 docker
```shell
curl -sSL https://resource.fit2cloud.com/1panel/package/quick_start.sh -o quick_start.sh && sudo bash quick_start.sh
```

## env 文件配置
- 配置文件示例:
```shell
SERVER_IP=192.168.1.100 # 服务器的ip地址
```

## 初始化数据库 (运行一次即可)
- 每次运行会删除数据库中的数据
```shell
sudo bash ./script/init-db.sh
```

## 1.初始化
```shell
sudo bash ./script/init.sh
```
## 2. 启动
```shell
sudo bash ./script/start.sh -r
```
## 注意点
- 每次运行`./script/init.sh`初始化脚本时,会进行环境初始化。
- 每次运行`./script/start.sh`启动脚本时，会启动所有容器，包括数据库容器。
- 每次运行`./script/init-db.sh`初始化数据库脚本时，会初始化数据库容器, 原来的数据将丢失。



