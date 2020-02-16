package main

import (
	"context"
	"flag"
	"fmt"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pb "github.com/ethanwu10/erebus/broker/gen"
)

var log *logrus.Logger

func run(port int) {
	netAddr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", netAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s", netAddr)
	}
	log.Printf("Listening on %s", netAddr)
	logrusEntry := logrus.NewEntry(log)
	// Shared options for the logger, with a custom gRPC code to log level function.
	opts := []grpc_logrus.Option{}
	grpc_logrus.ReplaceGrpcLogger(logrusEntry)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_logrus.UnaryServerInterceptor(logrusEntry, opts...),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_logrus.StreamServerInterceptor(logrusEntry, opts...),
		)),
	)
	broker := NewBroker(context.Background(), SimInfo{
		timestep: 32,
	})
	pb.RegisterWbControllerServer(server, NewWbControllerServer(broker))
	pb.RegisterClientControllerServer(server, NewClientControllerServer(broker))
	pb.RegisterControlServer(server, NewControlServer(broker))
	server.Serve(lis)
}

func main() {
	log = logrus.New()
	port := flag.Int("port", 51512, "port to listen on")
	flag.Parse()

	log.SetLevel(logrus.DebugLevel)

	run(*port)
}
