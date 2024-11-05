安装 zk kafka redis 

make build 
make run 
修改websocket地址


examples/javascript/client.js

var ws = new WebSocket('ws://10.66.0.11:3102/sub');


go run examples/javascript/main.go 


http://10.66.0.11:9080/ui/clusters/local/all-topics/goim-push-topic

http://10.66.0.11:1999/examples/javascript/