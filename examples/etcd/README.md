# etcd grpc

## 生成proto文件 greeter.proto 文件所在目录执行
protoc -I=. --go_out=. --go-grpc_out=. greeter.proto 