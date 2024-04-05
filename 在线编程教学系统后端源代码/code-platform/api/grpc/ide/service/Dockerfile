FROM lgbgbl/docker-dind

WORKDIR /app
COPY ./server /app/server
COPY ./pack.sh /app/pack.sh

CMD ["sh", "./pack.sh"]

