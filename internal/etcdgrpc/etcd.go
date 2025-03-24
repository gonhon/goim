package etcdgrpc

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/golang/glog"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	gresolver "google.golang.org/grpc/resolver"
)

const (
	nameServicePrefix = "goim"
	LocalDefName      = "rpc"
	LogicServerName   = "logic"
	CometServerName   = "comet"
	// 默认的租赁时间
	defLeaseSecond = 30
)

// 保存已经注册的服务
var serviceMap map[string]*EndpointUnit

// 服务元数据信息
type Endpoint struct {
	Name    string `json:"name"`
	Addr    string `json:"addr"`
	Port    int
	Version string `json:"version"`
}

// 服务元数据信息
type EndpointUnit struct {
	Name    string `json:"name"`
	LeaseID clientv3.LeaseID
	E       Endpoint
}

// 命名服务结构体
type NamingService struct {
	EtcdUrl string
	Name    string
	MTarget string
	Client  *clientv3.Client
	manager endpoints.Manager
}

// 创建一个新的命名服务
func NewNamingService(protocol, etcdHost string, etcdPort int, serviceName string) (*NamingService, error) {
	url := fmt.Sprintf("%s://%s:%d", protocol, etcdHost, etcdPort)
	client, err := clientv3.NewFromURL(url)
	if err != nil {
		return nil, err
	}
	target := fmt.Sprintf("%s/%s", nameServicePrefix, serviceName)
	log.Infof("target:%s\n", target)
	// etcd的endpoints管理
	manager, err := endpoints.NewManager(client, target)
	if err != nil {
		return nil, err
	}
	ns := NamingService{
		EtcdUrl: url,
		Name:    serviceName,
		MTarget: target,
		manager: manager,
		Client:  client,
	}
	serviceMap = make(map[string]*EndpointUnit)
	return &ns, nil
}

// 从本地etcd创建
func NewLocalDefNamingService(serviceName string) (*NamingService, error) {
	return NewNamingService("http", "localhost", 2379, serviceName)
}

func (naming *NamingService) GetFullServerName(name string, leaseID clientv3.LeaseID) string {
	return fmt.Sprintf("%s/%s/%d", naming.MTarget, name, leaseID)
}

func (naming *NamingService) GetPathServerName(name string) string {
	return fmt.Sprintf("%s/%s", naming.MTarget, name)
}

// 添加/注册新的服务
func (naming *NamingService) AddEndpoint(e Endpoint) error {
	b, _ := json.Marshal(e)
	ep := endpoints.Endpoint{
		Addr:     fmt.Sprintf("%s:%d", e.Addr, e.Port),
		Metadata: string(b), // 这里有个坑，必须传入字符串，没来得及看原因
	}
	// 在etcd创建一个续期的lease对象
	lease, err := naming.Client.Grant(context.TODO(), defLeaseSecond)
	if err != nil {
		return err
	}
	key := naming.GetFullServerName(e.Name, lease.ID)
	// 向etcd注册一个Endpoint并绑定续期
	err = naming.manager.AddEndpoint(context.TODO(), key, ep, clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}
	log.Info("AddEndpoint success:", key)
	// 开启自动续期KeepAlive
	ch, err := naming.Client.KeepAlive(context.TODO(), lease.ID)
	if err != nil {
		return err
	}
	// 这个方法会异步打印出每次续期调用的日志
	go func() {
		for {
			ka := <-ch
			log.Infof("ttl:", ka.ID, ka.TTL)
		}
	}()
	serviceMap[e.Name] = &EndpointUnit{Name: e.Name, LeaseID: lease.ID, E: e}
	return nil
}

// 移除一个服务
func (naming *NamingService) DelEndpoint(name string) error {
	eu := serviceMap[name]
	if eu == nil {
		return nil
	}
	err := naming.manager.DeleteEndpoint(context.TODO(), naming.GetFullServerName(name, eu.LeaseID))
	if err != nil {
		log.Fatalf("Delete endpoint error %v", err)
		return err
	}
	_, err = naming.Client.Revoke(context.TODO(), eu.LeaseID)
	if err != nil {
		log.Fatalf("Revoke lease error %v", err)
		return err
	}
	delete(serviceMap, name)
	log.Infof("DeleteEndpoint [%s] success\n", name)
	return nil
}

// 移除所有服务
func (naming *NamingService) DelAllEndpoint() {
	for k := range serviceMap {
		err := naming.DelEndpoint(k)
		if err != nil {
			log.Fatalln("Ignore Failure Continue...")
		}
	}
}

// 创建一个etcdResolver用于客户端发现服务
func (naming *NamingService) NewEtcdResolver() (gresolver.Builder, error) {
	etcdResolver, err := resolver.NewBuilder(naming.Client)
	if err != nil {
		log.Fatalf("Etcd resolver error %v", err)
		return nil, err
	}
	return etcdResolver, nil
}
