package comet

import (
	"context"
	"math/rand"
	"time"

	"github.com/Terry-Mao/goim/api/logic"
	"github.com/Terry-Mao/goim/internal/comet/conf"
	"github.com/Terry-Mao/goim/internal/etcdgrpc"
	log "github.com/golang/glog"
	"github.com/zhenjl/cityhash"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	minServerHeartbeat = time.Minute * 10
	maxServerHeartbeat = time.Minute * 30
	// grpc options
	grpcInitialWindowSize     = 1 << 24
	grpcInitialConnWindowSize = 1 << 24
	grpcMaxSendMsgSize        = 1 << 24
	grpcMaxCallMsgSize        = 1 << 24
	grpcKeepAliveTime         = time.Second * 10
	grpcKeepAliveTimeout      = time.Second * 3
	grpcBackoffMaxDelay       = time.Second * 3
)

func newLogicClient(c *conf.RPCClient) logic.LogicClient {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Dial))
	defer cancel()

	//注册etcd--开始
	service, err := etcdgrpc.NewLocalDefNamingService(etcdgrpc.LocalDefName)
	if err != nil {
		log.Fatalf("Create naming service error: %v", err)
	}

	resolver, err := service.NewEtcdResolver()
	if err != nil {
		log.Fatalf("Create etcd resolver error: %v", err)
	}

	target := "etcd://localhost:2379/" + service.GetPathServerName(etcdgrpc.LogicServerName)
	log.Infof("grpc.DialContext target:%s", target)

	conn, err := grpc.DialContext(ctx, target,
		[]grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithInitialWindowSize(grpcInitialWindowSize),
			grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
			grpc.WithBackoffMaxDelay(grpcBackoffMaxDelay),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                grpcKeepAliveTime,
				Timeout:             grpcKeepAliveTimeout,
				PermitWithoutStream: true,
			}),
			//负载均衡
			grpc.WithResolvers(resolver),
			grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		}...)

	if err != nil {
		panic(err)
	}
	/* etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}
	// defer etcdClient.Close()

	// 获取服务地址
	addresses, err := getServiceAddresses(etcdClient, "goim/logic")
	if err != nil {
		log.Fatalf("Failed to get service addresses: %v", err)
	}
	if len(addresses) == 0 {
		log.Fatalf("No service found")
	}

	// 随机选择一个服务地址进行请求
	// rand.Seed(time.Now().UnixNano())
	serviceAddress := addresses[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(addresses))]
	log.Infof("grpc serviceAddress:%s\n", serviceAddress)

	// 连接到 gRPC 服务
	conn, err := grpc.Dial(serviceAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	} */
	// defer conn.Close()
	//注册etcd--结束
	return logic.NewLogicClient(conn)
}

func getServiceAddresses(etcdClient *clientv3.Client, serviceName string) ([]string, error) {
	resp, err := etcdClient.Get(context.Background(), serviceName, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	var addresses []string
	for _, kv := range resp.Kvs {
		addresses = append(addresses, string(kv.Value))
	}
	return addresses, nil
}

// Server is comet server.
type Server struct {
	c         *conf.Config
	round     *Round    // accept round store
	buckets   []*Bucket // subkey bucket
	bucketIdx uint32

	serverID  string
	rpcClient logic.LogicClient
}

// NewServer returns a new Server.
func NewServer(c *conf.Config) *Server {
	s := &Server{
		c:         c,
		round:     NewRound(c),
		rpcClient: newLogicClient(c.RPCClient),
	}
	// init bucket
	s.buckets = make([]*Bucket, c.Bucket.Size)
	s.bucketIdx = uint32(c.Bucket.Size)
	for i := 0; i < c.Bucket.Size; i++ {
		s.buckets[i] = NewBucket(c.Bucket)
	}
	s.serverID = c.Env.Host
	go s.onlineproc()
	return s
}

// Buckets return all buckets.
func (s *Server) Buckets() []*Bucket {
	return s.buckets
}

// Bucket get the bucket by subkey.
func (s *Server) Bucket(subKey string) *Bucket {
	idx := cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % s.bucketIdx
	if conf.Conf.Debug {
		log.Infof("%s hit channel bucket index: %d use cityhash", subKey, idx)
	}
	return s.buckets[idx]
}

// RandServerHearbeat rand server heartbeat.
func (s *Server) RandServerHearbeat() time.Duration {
	return (minServerHeartbeat + time.Duration(rand.Int63n(int64(maxServerHeartbeat-minServerHeartbeat))))
}

// Close close the server.
func (s *Server) Close() (err error) {
	return
}

func (s *Server) onlineproc() {
	for {
		var (
			allRoomsCount map[string]int32
			err           error
		)
		roomCount := make(map[string]int32)
		for _, bucket := range s.buckets {
			for roomID, count := range bucket.RoomsCount() {
				roomCount[roomID] += count
			}
		}
		if allRoomsCount, err = s.RenewOnline(context.Background(), s.serverID, roomCount); err != nil {
			time.Sleep(time.Second)
			continue
		}
		for _, bucket := range s.buckets {
			bucket.UpRoomsCount(allRoomsCount)
		}
		time.Sleep(time.Second * 10)
	}
}
