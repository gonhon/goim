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