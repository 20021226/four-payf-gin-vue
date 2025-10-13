FROM alpine:latest

COPY . /app
WORKDIR /app

# 设置时区
ENV TZ=Asia/Shanghai
RUN apk update && apk add --no-cache tzdata openntpd \
    && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

EXPOSE 8888
ENTRYPOINT ./server -c config.docker.yaml
