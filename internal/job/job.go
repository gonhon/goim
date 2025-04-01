package job

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	pb "github.com/Terry-Mao/goim/api/logic"
	"github.com/Terry-Mao/goim/internal/etcdgrpc"
	"github.com/Terry-Mao/goim/internal/job/conf"
	"github.com/golang/protobuf/proto"
	clientv3 "go.etcd.io/etcd/client/v3"

	cluster "github.com/bsm/sarama-cluster"
	log "github.com/golang/glog"
)

// Job is push job.
type Job struct {
	c            *conf.Config
	consumer     *cluster.Consumer
	cometServers map[string]*Comet

	rooms      map[string]*Room
	roomsMutex sync.RWMutex
}

// New new a push job.
func New(c *conf.Config) *Job {
	j := &Job{
		c:        c,
		consumer: newKafkaSub(c.Kafka),
		rooms:    make(map[string]*Room),
	}
	// j.watchComet(c.Discovery)
	j.watchCometEtcd()
	return j
}

func newKafkaSub(c *conf.Kafka) *cluster.Consumer {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	consumer, err := cluster.NewConsumer(c.Brokers, c.Group, []string{c.Topic}, config)
	if err != nil {
		panic(err)
	}
	return consumer
}

// Close close resounces.
func (j *Job) Close() error {
	if j.consumer != nil {
		return j.consumer.Close()
	}
	return nil
}

// Consume messages, watch signals
func (j *Job) Consume() {
	for {
		select {
		case err := <-j.consumer.Errors():
			log.Errorf("consumer error(%v)", err)
		case n := <-j.consumer.Notifications():
			log.Infof("consumer rebalanced(%v)", n)
		case msg, ok := <-j.consumer.Messages():
			if !ok {
				return
			}
			j.consumer.MarkOffset(msg, "")
			// process push message
			pushMsg := new(pb.PushMsg)
			if err := proto.Unmarshal(msg.Value, pushMsg); err != nil {
				log.Errorf("proto.Unmarshal(%v) error(%v)", msg, err)
				continue
			}
			if err := j.push(context.Background(), pushMsg); err != nil {
				log.Errorf("j.push(%v) error(%v)", pushMsg, err)
			}
			log.Infof("consume: %s/%d/%d\t%s\t%+v", msg.Topic, msg.Partition, msg.Offset, msg.Key, pushMsg)
		}
	}
}
func (j *Job) watchCometEtcd() {
	service, err := etcdgrpc.NewLocalDefNamingService(etcdgrpc.LocalRpcName)
	if err != nil {
		panic(err)
	}

	watchKey := fmt.Sprintf("%s/%s/%s", etcdgrpc.NameServicePrefix, etcdgrpc.LocalDataName, etcdgrpc.CometServerName)
	log.Infof("watch key:%s", watchKey)

	resp, err := service.Client.Get(context.Background(), watchKey, clientv3.WithPrefix())
	if err == nil && len(resp.Kvs) > 0 {
		for _, v := range resp.Kvs {
			ins := &etcdgrpc.Instance{}
			json.Unmarshal(v.Value, ins)
			j.newAddressEtcd([]*etcdgrpc.Instance{ins})
		}
	}
	go func() {
		watchChan := service.Client.Watch(context.Background(), watchKey, clientv3.WithPrefix())
		for watch := range watchChan {
			for _, event := range watch.Events {
				switch event.Type {
				case clientv3.EventTypePut:
					log.Infof("Key updated: %s, Value: %s\n", event.Kv.Key, event.Kv.Value)
					ins := &etcdgrpc.Instance{}
					json.Unmarshal(event.Kv.Value, ins)
					j.newAddressEtcd([]*etcdgrpc.Instance{ins})
				case clientv3.EventTypeDelete:
					log.Infof("Key deleted: %s\n", event.Kv.Key)
				}
			}
		}
	}()
}

func (j *Job) newAddressEtcd(insArr []*etcdgrpc.Instance) error {
	comets := map[string]*Comet{}
	for _, in := range insArr {
		if old, ok := j.cometServers[in.Hostname]; ok {
			comets[in.Hostname] = old
			continue
		}
		c, err := NewComet(in, j.c.Comet)
		if err != nil {
			log.Errorf("watchComet NewComet(%+v) error(%v)", in, err)
			return err
		}
		comets[in.Hostname] = c
		log.Infof("watchComet AddComet grpc:%+v", in)
	}
	for key, old := range j.cometServers {
		if _, ok := comets[key]; !ok {
			old.cancel()
			log.Infof("watchComet DelComet:%s", key)
		}
	}
	j.cometServers = comets
	return nil
}
