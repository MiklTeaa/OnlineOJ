# 自定义 docker dind镜像说明
改自 https://github.com/docker-library/docker/blob/master/20.10/dind  
1. 抹除了 EXPOSE 属性，避免多个 docker:dind 镜像同时使用 host 模式部署时，发生端口占用而无法启动的现象。
2. 直接使用 dockerd 运行 docker daemon，启动更快。

打包命令：docker build --network=host -t lgbgbl/docker-dind -f Dockerfile .