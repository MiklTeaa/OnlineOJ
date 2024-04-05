protoc --go_out=plugins=grpc:./../../../api/grpc/ide  ./ide.proto

protoc --go_out=plugins=grpc:./../../../api/grpc/monaco  ./monaco.proto

protoc --go_out=plugins=grpc:./../../../service/user  ./user.proto