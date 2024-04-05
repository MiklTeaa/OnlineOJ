GOOS=linux go build -o ./api/grpc/ide/service/server  ./api/grpc/ide/service/*.go
GOOS=linux go build -o ./api/grpc/monaco/service/server  ./api/grpc/monaco/service/*.go

cd dev
docker-compose --compatibility up -d --build

cd -
rm ./api/grpc/ide/service/server ./api/grpc/monaco/service/server
