FROM alpine:latest

COPY . /app
WORKDIR /app

# 设置时区
ENV TZ=Asia/Shanghai
# 配置阿里云镜像源 + 时区设置 + 安装依赖
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk update \
    && apk add --no-cache tzdata openntpd \
    && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime \
    && echo $TZ > /etc/timezone \
    && rm -rf /var/cache/apk/*

EXPOSE 8888
ENTRYPOINT /app/server -c config.release.yaml
