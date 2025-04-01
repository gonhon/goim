package grpc

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/golang/glog"

	pb "github.com/Terry-Mao/goim/api/logic"
	"github.com/Terry-Mao/goim/internal/etcdgrpc"
	"github.com/Terry-Mao/goim/internal/logic"
	"github.com/Terry-Mao/goim/internal/logic/conf"
	clientv3 "go.etcd.io/etcd/client/v3"

	"google.golang.org/grpc"

	// use gzip decoder
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// New logic grpc server
func New(c *conf.RPCServer, l *logic.Logic) *grpc.Server {
	keepParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(c.IdleTimeout),
		MaxConnectionAgeGrace: time.Duration(c.ForceCloseWait),
		Time:                  time.Duration(c.KeepAliveInterval),
		Timeout:               time.Duration(c.KeepAliveTimeout),
		MaxConnectionAge:      time.Duration(c.MaxLifeTime),
	})
	srv := grpc.NewServer(keepParams)
	// srv := grpc.NewServer()
	pb.RegisterLogicServer(srv, &server{l})
	// 在服务注册后添加反射
	reflection.Register(srv)
	//注册etcd--开始
	grpcPort, _ := strconv.Atoi(strings.TrimPrefix(c.Addr, ":"))
	log.Infof("%s gprc port %d", etcdgrpc.LogicServerName, grpcPort)
	err := l.NamingService.AddEndpoint(etcdgrpc.Endpoint{
		Addr:    "localhost",
		Name:    etcdgrpc.LogicServerName,
		Port:    grpcPort,
		Version: "1.0.0",
	})
	if err != nil {
		panic(err)
	}
	//注册etcd--结束

	lis, err := net.Listen(c.Network, c.Addr)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := srv.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return srv
}

func registerService(client *clientv3.Client, serviceName, serviceAddr string) {
	lease, err := client.Grant(context.Background(), 60) // 租约10秒
	if err != nil {
		log.Fatalf("Failed to grant lease: %v", err)
	}

	instanceID := fmt.Sprintf("%s/%s", serviceName, serviceAddr) // 使用唯一的 key
	log.Infof("instanceID:%s\n", instanceID)
	_, err = client.Put(context.Background(), instanceID, serviceAddr, clientv3.WithLease(lease.ID))
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 定期续租
	/* go func() {
		for {
			_, err = client.KeepAlive(context.Background(), lease.ID)
			if err != nil {
				log.Fatalf("Failed to keep alive: %v", err)
			}
			time.Sleep(5 * time.Second)
		}
	}() */
	// 开始续租
	keepAliveCh, err := client.KeepAlive(context.Background(), lease.ID)
	if err != nil {
		log.Fatalf("Failed to start keep alive: %v", err)
	}

	go func() {
		for {
			select {
			case ka, ok := <-keepAliveCh:
				if !ok {
					log.Infof("Failed to keep alive: %v", err)
					return
				}
				log.Infof("Keep alive response: %v", ka)
			}
		}
	}()
	/* go func() {
		for {
			_, err := client.KeepAliveOnce(context.Background(), lease.ID)
			if err != nil {
				log.Infof("Failed to keep alive: %v", err)
				return
			}
			time.Sleep(5 * time.Second) // 设置心跳间隔
		}
	}() */
}

type server struct {
	srv *logic.Logic
}

var _ pb.LogicServer = &server{}

// Connect connect a conn.
func (s *server) Connect(ctx context.Context, req *pb.ConnectReq) (*pb.ConnectReply, error) {
	mid, key, room, accepts, hb, err := s.srv.Connect(ctx, req.Server, req.Cookie, req.Token)
	if err != nil {
		return &pb.ConnectReply{}, err
	}
	return &pb.ConnectReply{Mid: mid, Key: key, RoomID: room, Accepts: accepts, Heartbeat: hb}, nil
}

// Disconnect disconnect a conn.
func (s *server) Disconnect(ctx context.Context, req *pb.DisconnectReq) (*pb.DisconnectReply, error) {
	has, err := s.srv.Disconnect(ctx, req.Mid, req.Key, req.Server)
	if err != nil {
		return &pb.DisconnectReply{}, err
	}
	return &pb.DisconnectReply{Has: has}, nil
}

// Heartbeat beartbeat a conn.
func (s *server) Heartbeat(ctx context.Context, req *pb.HeartbeatReq) (*pb.HeartbeatReply, error) {
	if err := s.srv.Heartbeat(ctx, req.Mid, req.Key, req.Server); err != nil {
		return &pb.HeartbeatReply{}, err
	}
	return &pb.HeartbeatReply{}, nil
}

// RenewOnline renew server online.
func (s *server) RenewOnline(ctx context.Context, req *pb.OnlineReq) (*pb.OnlineReply, error) {
	allRoomCount, err := s.srv.RenewOnline(ctx, req.Server, req.RoomCount)
	if err != nil {
		return &pb.OnlineReply{}, err
	}
	return &pb.OnlineReply{AllRoomCount: allRoomCount}, nil
}

// Receive receive a message.
func (s *server) Receive(ctx context.Context, req *pb.ReceiveReq) (*pb.ReceiveReply, error) {
	if err := s.srv.Receive(ctx, req.Mid, req.Proto); err != nil {
		return &pb.ReceiveReply{}, err
	}
	return &pb.ReceiveReply{}, nil
}

// nodes return nodes.
func (s *server) Nodes(ctx context.Context, req *pb.NodesReq) (*pb.NodesReply, error) {
	return s.srv.NodesWeighted(ctx, req.Platform, req.ClientIP), nil
}
