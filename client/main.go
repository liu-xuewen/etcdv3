package main

import (
	"google.golang.org/grpc/resolver"
	"etcdv3/client/etcd"
	"etcdv3/client/pb"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"time"
	"context"
)

var (
	etcdLoc = "localhost:2379"
)

func main() {
	hiBuilder := etcd.NewResolverBuilder([]string{etcdLoc})
	resolver.Register(&hiBuilder)
	var dialAddr = fmt.Sprintf("%s://foo/%s", hiBuilder.Scheme(), "serverTest")
	conn, err := grpc.Dial(dialAddr, grpc.WithBalancerName("round_robin"), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewServerClient(conn)
	for range time.Tick(time.Second * 3) {
		req := &pb.TestReq{Name: "Foo"}
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
		defer cancelFunc()
		resp, err := client.Test(ctx, req)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v \n", resp)
	}
}