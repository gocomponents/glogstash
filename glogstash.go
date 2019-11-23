package main

import (
	"context"
	"github.com/gocomponents/core/proto"
	"github.com/gocomponents/glogstash/produce_consume"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

func main()  {
	go produce_consume.Consume()
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()

	proto.RegisterLogStashServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		panic(err)
	}
}

type server struct{}

func (s *server) Send(ctx context.Context, request *proto.Log) (*proto.Response, error) {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("post to channel panic,%v", err)
		}
	}()

	go produce_consume.Produce(request)

	return &proto.Response{
		ErrorCode: 0,
		Message:   "",
	}, nil
}

