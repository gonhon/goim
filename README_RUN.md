安装 zk kafka redis 

make build 
make run 
修改websocket地址


examples/javascript/client.js

var ws = new WebSocket('ws://10.66.0.11:3102/sub');


go run examples/javascript/main.go 


http://10.66.0.11:9080/ui/clusters/local/all-topics/goim-push-topic

http://10.66.0.11:1999/examples/javascript/

debug
```json
 "args": [
                "-region","sh",
                "-zone","sh001",
                "-deploy.env","dev",
            ]
```

```
# Makefile

# 交叉编译为MacOS
build-macos:
	GOOS=darwin GOARCH=amd64 go build -o bin/ssh-tunnel-app

# 交叉编译为Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/ssh-tunnel-app

# 交叉编译为Windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/ssh-tunnel-app.exe

# 默认目标，执行全部编译
all: build-macos build-linux build-windows
```



在 GoIM（一个用于即时通讯的 Go 语言框架）中，`bucket`、`channel`、`operation`、`ring`、`room` 和 `round` 各自有不同的作用。以下是这些概念的简要说明：

### 1. Bucket
- **作用**：Bucket 是用来存储用户会话的容器。它通常用于管理连接的用户和他们的会话信息。
- **用途**：可以将用户的连接信息进行分组，以便高效地进行消息路由和状态管理。

### 2. Channel
- **作用**：Channel 是用于消息传递的通道。它可以用于在不同的 goroutine 之间进行通信。
- **用途**：在 GoIM 中，channel 可以帮助实现消息的发送和接收，确保消息在并发环境下的安全传递。

### 3. Operation
- **作用**：Operation 指的是对消息或事件的处理操作。每个操作通常对应一个具体的功能，如发送消息、用户登录、用户退出等。
- **用途**：通过定义不同的操作，GoIM 可以灵活地处理多种类型的请求和事件。

### 4. Ring
- **作用**：Ring 是一种数据结构，用于管理在线用户的状态，通常实现为循环队列。
- **用途**：它可以帮助高效地存储和检索在线用户的信息，并在用户状态变化时进行更新。

### 5. Room
- **作用**：Room 表示一个聊天房间，用户可以在其中进行交流。每个房间可以包含多个用户。
- **用途**：通过房间的概念，可以实现群聊功能，使得用户能够在同一个房间中发送和接收消息。

### 6. Round
- **作用**：Round 通常指的是一次会话或交互的周期。它可以用来管理消息的发送和接收过程。
- **用途**：在某些场景下，Round 可以帮助跟踪消息的状态或处理逻辑，以确保消息按顺序处理。

### 总结
这些概念共同构成了 GoIM 的基础架构，使其能够高效、可靠地处理即时通讯的需求。每个组件都有其特定的功能，互相配合以实现完整的消息传递和用户管理功能。


TEST

http://10.66.0.11/im/mids1/
http://10.66.0.11/im/mids2/
http://10.66.0.11/im/mids123_room2/
http://10.66.0.11:9080/ui/clusters/local/all-topics/goim-push-topic


I1221 16:11:03.679344   11477 server_websocket.go:145] start tcp serve "10.66.0.11:3102" with "10.66.0.1:58969"
E1221 16:11:06.907513   11477 server_websocket.go:226] key:  remoteIP: 10.66.0.1:58969 step: 4 ws handshake failed error(rpc error: code = Unavailable desc = all SubConns are in TransientFailure, latest connection error: connection closed)
I1221 16:11:10.110675   11477 client.go:458] discovery: set client.Get(http://127.0.0.1:7171/discovery/set?appid=goim.comet&env=dev&hostname=dev-server&metadata=%7B%22addrs%22%3A%22127.0.0.1%22%2C%22conn_count%22%3A%220%22%2C%22ip_count%22%3A%220%22%2C%22offline%22%3A%22false%22%2C%22weight%22%3A%2210%22%7D&region=sh&status=1&version=&zone=sh001) env(dev) appid(goim.comet) addrs([grpc://172.19.134.80:3109]) success
I1221 16:11:20.112775   11477 client.go:458] discovery: set client.Get(http://127.0.0.1:7171/discovery/set?appid=goim.comet&env=dev&hostname=dev-server&metadata=%7B%22addrs%22%3A%22127.0.0.1%22%2C%22conn_count%22%3A%220%22%2C%22ip_count%22%3A%220%22%2C%22offline%22%3A%22false%22%2C%22weight%22%3A%2210%22%7D&region=sh&status=1&version=&zone=sh001) env(dev) appid(goim.comet) addrs([grpc://172.19.134.80:3109]) success
I1221 16:11:20.113586   11477 client.go:569] discovery: successfully polls(http://127.0.0.1:7171/discovery/polls?appid=infra.discovery&appid=goim.logic&env=dev&hostname=dev-server&latest_timestamp=1734768270630885245&latest_timestamp=1734768511101364041) instances ({"goim.comet":{"instances":{"sh001":[{"region":"sh","zone":"sh001","env":"dev","appid":"goim.comet","hostname":"dev-server","addrs":["grpc://172.19.134.80:3109"],"version":"","latest_timestamp":0,"metadata":{"addrs":"127.0.0.1","conn_count":"0","ip_count":"0","offline":"false","weight":"10"}}]},"latest_timestamp":1734768510061289243,"scheduler":null}})


google.golang.org/grpc
github.com/bilibili/discovery/naming
github.com/bsm/sarama-cluster


go env -w GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct

go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOPROXY=https://proxy.golang.org,direct

go clean -modcache && go mod tidy

go get go.etcd.io/etcd/client/v3


root@dev-server:/resources/codes/go/goim# make build 
rm -rf target/
mkdir target/
cp cmd/comet/comet-example.toml target/comet.toml
cp cmd/logic/logic-example.toml target/logic.toml
cp cmd/job/job-example.toml target/job.toml
GO111MODULE=on go build -gcflags="all=-N -l" -o target/comet cmd/comet/main.go
# github.com/bilibili/discovery/naming/grpc
/usr/local/go/pkg/mod/github.com/bilibili/discovery@v1.0.1/naming/grpc/resolver.go:44:87: undefined: resolver.BuildOption
/usr/local/go/pkg/mod/github.com/bilibili/discovery@v1.0.1/naming/grpc/resolver.go:46:24: cannot use target.Endpoint (value of type func() string) as string value in argument to strings.SplitN
/usr/local/go/pkg/mod/github.com/bilibili/discovery@v1.0.1/naming/grpc/resolver.go:95:42: undefined: resolver.ResolveNowOption
/usr/local/go/pkg/mod/github.com/bilibili/discovery@v1.0.1/naming/grpc/resolver.go:150:4: unknown field Type in struct literal of type "google.golang.org/grpc/resolver".Address
/usr/local/go/pkg/mod/github.com/bilibili/discovery@v1.0.1/naming/grpc/resolver.go:150:25: undefined: resolver.Backend
# github.com/Terry-Mao/goim/internal/comet
internal/comet/server.go:46:9: undefined: grpc.WithBalancerName
make: *** [Makefile:13: build] Error 1




protoc --go_out=. --go-grpc_out=. comet.proto
protoc --proto_path=/resources/codes/go/goim/api/comet --proto_path=/resources/codes/go/goim/api/protocol --go_out=. --go-grpc_out=. /resources/codes/go/goim/api/comet/comet_new.proto
protoc --proto_path=../../ --go_out=. --go-grpc_out=. comet_new.proto