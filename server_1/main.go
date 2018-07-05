package main

import (
	"context"
	"etcdv3/server_1/pb"
	"net"
	"fmt"
	"log"
	"etcdv3/server_1/etcdv3"
	"google.golang.org/grpc"
)

const (
	srvName = "serverTest"
	port    = ":10013"
)

type Server struct {

}

func (s *Server)Test(ctx context.Context, req *pb.TestReq) (*pb.TestResp, error) {
	return &pb.TestResp{
		Echo : "    Hi " + req.Name + ", this response comes from Server 10013",
	},nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	reg := etcdv3.NewReg([]string{"127.0.0.1:2379"}, "nxf")
	err = reg.Register(srvName, fmt.Sprintf("127.0.0.1%s", port))
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	pb.RegisterServerServer(s, &Server{})

	fmt.Println("Registger...")

	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}