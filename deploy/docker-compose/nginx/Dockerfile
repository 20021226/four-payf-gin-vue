# 如果需要用 cicd ，请设置环境变量：
# variables:
#     DOCKER_BUILDKIT: 1
FROM nginx:alpine

COPY ./conf.d/ /etc/nginx/conf.d
COPY ./html/dist /usr/share/nginx/html
COPY ./ssl /etc/nginx/ssl
RUN ls -al /usr/share/nginx/html
