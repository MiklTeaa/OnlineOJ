# gen rpc stub files
protoc --java_out ../../../api/grpc/plagiarismDetection/service/src/main/java ./*.proto

# gen grpc server files according to proto files
protoc --grpc-java_out=../../../api/grpc/plagiarismDetection/service/src/main/java ./*.proto


protoc --go_out=plugins=grpc:./../../../api/grpc/plagiarismDetection  ./*.proto
