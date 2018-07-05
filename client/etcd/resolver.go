package etcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"google.golang.org/grpc/resolver"
	"log"
	"strings"
	"time"
	"runtime/debug"
)

type Resolver struct {
}

func (r *Resolver) Close() {
	fmt.Println(" close method ")
}

func (r *Resolver) ResolveNow(opt resolver.ResolveNowOption) {
	debug.PrintStack()
	fmt.Println(" ResolveNow method ")
}

func NewResolverBuilder(etcdEndPoints []string) Builder {
	return Builder{
		endPoints: etcdEndPoints,
	}
}

type Builder struct {
	cc        resolver.ClientConn
	endPoints []string
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	var err error
	if cli == nil {
		cli, err = clientv3.New(clientv3.Config{
			Endpoints:   b.endPoints,
			DialTimeout: 10 * time.Second,
		})
		if err != nil {
			return nil, err
		}
	}
	b.cc = cc
	r := &Resolver{}
	go b.watcher(fmt.Sprintf("%s/%s/", target.Scheme, target.Endpoint))
	return r, nil
}

func (b *Builder) watcher(keyPrefix string) {
	var addrList []resolver.Address
	//第一次拉取所有的地址
	resp, err := cli.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}
	for _, kv := range resp.Kvs {
		addr := resolver.Address{
			Addr: strings.Trim(string(kv.Key), keyPrefix),
		}
		addrList = append(addrList, addr)
	}
	//重新设置地址给grpc clientConn ，clientconn 监听chan(grpc/resolver_conn_wrapper.go , line 142)
	b.cc.NewAddress(addrList)
	watch := cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	for wresp := range watch {
		for _, ev := range wresp.Events {
			evKey := strings.TrimPrefix(string(ev.Kv.Key), keyPrefix)
			switch ev.Type {
			case mvccpb.PUT:
				if !exist(addrList, evKey) {
					addrList = append(addrList, resolver.Address{Addr: evKey})
					b.cc.NewAddress(addrList)
				}
			case mvccpb.DELETE:
				if list, ok := remove(addrList, evKey); ok {
					addrList = list
					b.cc.NewAddress(addrList)
				}
			}
		}
	}
}

func (b *Builder) Scheme() string {
	return DefaultScheme
}

func exist(l []resolver.Address, addr string) bool {
	for i := range l {
		if l[i].Addr == addr {
			return true
		}
	}
	return false
}

func remove(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}