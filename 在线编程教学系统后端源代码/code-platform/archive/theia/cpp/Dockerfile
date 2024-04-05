ARG NODE_VERSION=12.18.3
FROM node:${NODE_VERSION}-slim

RUN  sed -i s@/deb.debian.org/@/mirrors.aliyun.com/@g /etc/apt/sources.list
RUN  apt-get clean


RUN apt-get update && \
    apt-get -y install libsecret-1-0 && \
    rm -rf /var/lib/apt/lists/*


RUN apt-get update && apt-get -y install build-essential libsecret-1-dev --fix-missing;
ARG version=latest
WORKDIR /home/theia
ADD $version.package.json ./package.json
RUN yarn --pure-lockfile && \
    NODE_OPTIONS="--max_old_space_size=4096" yarn theia build && \
    yarn theia download:plugins && \
    yarn --production && \
    yarn autoclean --init && \
    echo *.ts >> .yarnclean && \
    echo *.ts.map >> .yarnclean && \
    echo *.spec.* >> .yarnclean && \
    yarn autoclean --force && \
    yarn cache clean

FROM node:${NODE_VERSION}-slim



COPY --from=0 /home/theia /home/theia
WORKDIR /home/theia

# C/C++ Developer tools
RUN  sed -i s@/deb.debian.org/@/mirrors.aliyun.com/@g /etc/apt/sources.list \
    && sed -i s@/security.debian.org/@/mirrors.aliyun.com/@g /etc/apt/sources.list
RUN  apt-get clean



RUN apt-get update 
RUN apt-get install -y g++ gcc
RUN apt-get install -y libsecret-1-dev
RUN apt-get install -y clangd-11
RUN ln -s /usr/bin/clangd-11 /usr/bin/clangd


RUN addgroup theia && \
    adduser --ingroup theia --shell /bin/sh theia;
RUN chmod g+rw /home && \
    mkdir -p /home/project && \
    chown -R theia:theia /home/theia && \
    chown -R theia:theia /home/project;


ENV HOME /home/theia
WORKDIR /home/theia
COPY --from=0 /home/theia /home/theia
EXPOSE 3000
ENV SHELL=/bin/bash \
    THEIA_DEFAULT_PLUGINS=local-dir:/home/theia/plugins

USER theia
ENTRYPOINT [ "node", "/home/theia/src-gen/backend/main.js", "/home/project", "--hostname=0.0.0.0" ]
