ARG theia
FROM $theia

USER root
RUN npm install -g gen-http-proxy

ADD pack.sh /home/theia/ssl/

ENV server :10443
ENV staticfolder /usr/local/lib/node_modules/gen-http-proxy/static

ENTRYPOINT [ "sh", "/home/theia/ssl/pack.sh" ]