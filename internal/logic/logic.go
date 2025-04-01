package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Terry-Mao/goim/internal/etcdgrpc"
	"github.com/Terry-Mao/goim/internal/logic/conf"
	"github.com/Terry-Mao/goim/internal/logic/dao"
	"github.com/Terry-Mao/goim/internal/logic/model"
	log "github.com/golang/glog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	_onlineTick     = time.Second * 10
	_onlineDeadline = time.Minute * 5
)

// Logic struct
type Logic struct {
	c *conf.Config
	// dis *naming.Discovery
	dao *dao.Dao
	// online
	totalIPs   int64
	totalConns int64
	roomCount  map[string]int32
	// load balancer
	nodes         []*etcdgrpc.Instance
	loadBalancer  *LoadBalancer
	regions       map[string]string // province -> region
	NamingService *etcdgrpc.NamingService
}

// New init
func New(c *conf.Config) (l *Logic) {
	l = &Logic{
		c:   c,
		dao: dao.New(c),
		// dis:          naming.New(c.Discovery),
		loadBalancer: NewLoadBalancer(),
		regions:      make(map[string]string),
	}
	//初始化etcd相关信息
	service, err := etcdgrpc.NewLocalDefNamingService(etcdgrpc.LocalRpcName)
	if err != nil {
		panic(err)
	}
	l.NamingService = service
	l.initRegions()
	l.initNodes()
	_ = l.loadOnline()
	go l.onlineproc()
	return l
}

// Ping ping resources is ok.
func (l *Logic) Ping(c context.Context) (err error) {
	return l.dao.Ping(c)
}

// Close close resources.
func (l *Logic) Close() {
	l.dao.Close()
}

func (l *Logic) initRegions() {
	for region, ps := range l.c.Regions {
		for _, province := range ps {
			l.regions[province] = region
		}
	}
}

func (l *Logic) initNodes() {
	cli := l.NamingService.Client
	watchKey := fmt.Sprintf("%s/%s/%s", etcdgrpc.NameServicePrefix, etcdgrpc.LocalDataName, etcdgrpc.CometServerName)
	log.Infof("watch key:%s", watchKey)
	resp, err := cli.Get(context.Background(), watchKey, clientv3.WithPrefix())
	if err == nil && len(resp.Kvs) > 0 {
		resArray := make([]*etcdgrpc.Instance, len(resp.Kvs))
		for _, v := range resp.Kvs {
			ins := &etcdgrpc.Instance{}
			json.Unmarshal(v.Value, ins)
			resArray = append(resArray, ins)
		}
		l.newNodesEtcd(resArray)
	}

	go func() {
		watchChan := cli.Watch(context.Background(), watchKey, clientv3.WithPrefix())
		for watchResp := range watchChan {
			for _, event := range watchResp.Events {
				switch event.Type {
				case clientv3.EventTypePut:
					log.Infof("Key updated: %s, Value: %s\n", event.Kv.Key, event.Kv.Value)
					ins := &etcdgrpc.Instance{}
					json.Unmarshal(event.Kv.Value, ins)
					l.newNodesEtcd([]*etcdgrpc.Instance{ins})
				case clientv3.EventTypeDelete:
					log.Infof("Key deleted: %s\n", event.Kv.Key)
				}
			}
		}
	}()

	/* res := l.dis.Build("goim.comet")
	event := res.Watch()
	select {
	case _, ok := <-event:
		if ok {
			l.newNodes(res)
		} else {
			panic("discovery watch failed")
		}
	case <-time.After(10 * time.Second):
		log.Error("discovery start timeout")
	}
	go func() {
		for {
			if _, ok := <-event; !ok {
				return
			}
			l.newNodes(res)
		}
	}() */
}

/* func (l *Logic) newNodes(res naming.Resolver) {
	if zoneIns, ok := res.Fetch(); ok {
		var (
			totalConns int64
			totalIPs   int64
			allIns     []*naming.Instance
		)
		for _, zins := range zoneIns.Instances {
			for _, ins := range zins {
				if ins.Metadata == nil {
					log.Errorf("node instance metadata is empty(%+v)", ins)
					continue
				}
				offline, err := strconv.ParseBool(ins.Metadata[model.MetaOffline])
				if err != nil || offline {
					log.Warningf("strconv.ParseBool(offline:%t) error(%v)", offline, err)
					continue
				}
				conns, err := strconv.ParseInt(ins.Metadata[model.MetaConnCount], 10, 32)
				if err != nil {
					log.Errorf("strconv.ParseInt(conns:%d) error(%v)", conns, err)
					continue
				}
				ips, err := strconv.ParseInt(ins.Metadata[model.MetaIPCount], 10, 32)
				if err != nil {
					log.Errorf("strconv.ParseInt(ips:%d) error(%v)", ips, err)
					continue
				}
				totalConns += conns
				totalIPs += ips
				allIns = append(allIns, ins)
			}
		}
		l.totalConns = totalConns
		l.totalIPs = totalIPs
		l.nodes = allIns
		l.loadBalancer.Update(allIns)
	}
} */

func (l *Logic) newNodesEtcd(res []*etcdgrpc.Instance) {
	var (
		totalConns int64
		totalIPs   int64
		allIns     []*etcdgrpc.Instance
	)
	for _, ins := range res {
		if ins == nil {
			log.Errorf("node instance is empty(%+v)", ins)
			continue
		}
		if ins.Metadata == nil {
			log.Errorf("node instance metadata is empty(%+v)", ins)
			continue
		}
		offline, err := strconv.ParseBool(ins.Metadata[model.MetaOffline])
		if err != nil || offline {
			log.Warningf("strconv.ParseBool(offline:%t) error(%v)", offline, err)
			continue
		}
		conns, err := strconv.ParseInt(ins.Metadata[model.MetaConnCount], 10, 32)
		if err != nil {
			log.Errorf("strconv.ParseInt(conns:%d) error(%v)", conns, err)
			continue
		}
		ips, err := strconv.ParseInt(ins.Metadata[model.MetaIPCount], 10, 32)
		if err != nil {
			log.Errorf("strconv.ParseInt(ips:%d) error(%v)", ips, err)
			continue
		}
		totalConns += conns
		totalIPs += ips
		allIns = append(allIns, ins)
	}
	l.totalConns = totalConns
	l.totalIPs = totalIPs
	l.nodes = allIns
	l.loadBalancer.Update(allIns)
}

func (l *Logic) onlineproc() {
	for {
		time.Sleep(_onlineTick)
		if err := l.loadOnline(); err != nil {
			log.Errorf("onlineproc error(%v)", err)
		}
	}
}

func (l *Logic) loadOnline() (err error) {
	var (
		roomCount = make(map[string]int32)
	)
	for _, server := range l.nodes {
		var online *model.Online
		online, err = l.dao.ServerOnline(context.Background(), server.Hostname)
		if err != nil {
			return
		}
		if time.Since(time.Unix(online.Updated, 0)) > _onlineDeadline {
			_ = l.dao.DelServerOnline(context.Background(), server.Hostname)
			continue
		}
		for roomID, count := range online.RoomCount {
			roomCount[roomID] += count
		}
	}
	l.roomCount = roomCount
	return
}
