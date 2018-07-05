package etcdv3

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

var (
	cli *clientv3.Client
)

const (
	DefaultScheme = "nxf"
)

type Reg struct {
	Endpoints []string
	Scheme    string
}

func NewReg(endpoints []string, scheme string) Reg {
	fmt.Println("222")
	if len(scheme) == 0 {
		scheme = DefaultScheme
	}
	return Reg{
		Endpoints: endpoints,
		Scheme:    scheme,
	}
}

func (r *Reg) Unregister(serverName string) error {
	_, err := cli.Delete(context.Background(), serverName, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	return nil
}

// Grant：分配一个租约
// Revoke：释放一个租约
// TimeToLive：获取剩余TTL时间
// Leases：列举所有etcd中的租约
// KeepAlive：自动定时的续约某个租约
// KeepAliveOnce：为某个租约续约一次
// Close：关闭当前客户端建立的所有租约
func (r *Reg) Register(name, addr string) error {
	var err error
	if cli == nil {
		cli, err = clientv3.New(clientv3.Config{
			Endpoints:   r.Endpoints,
			DialTimeout: 10 * time.Second,
		})
		if err != nil {
			return err
		}
	}
	//实现key自动过期 必须要创建一个租约 ，可设置ttl
	leaseResp, err := cli.Grant(context.Background(), 12)
	//主要使用到了租约id
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s/%s/%s", r.Scheme, name, addr)
	_, err = cli.Put(context.Background(), key, addr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	//当我们实现服务注册时，需要主动给Lease进行续约，这需要调用KeepAlive/KeepAliveOnce，你可以在一个循环中定时的调用：
	//需要注意 KeepAlive和Put一样，如果在执行之前Lease就已经过期了，那么需要重新分配Lease
	//Etcd并没有提供API来实现原子的Put with Lease ，必须重新grant

	go func() {
		for range time.Tick(time.Second * 10) {
			cli.KeepAliveOnce(context.Background(), leaseResp.ID)
		}
	}()
	return nil
}