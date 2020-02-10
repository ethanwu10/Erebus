package cmd

import (
	"fmt"
	"os"

	"google.golang.org/grpc"

	pb "github.com/ethanwu10/erebus/broker-control-cli/gen"
)

func getControlClient() pb.ControlClient {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to server \"%s\"\n", server)
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	return pb.NewControlClient(conn)
}
