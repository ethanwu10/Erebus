package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pb "github.com/ethanwu10/erebus/broker-control-cli/gen"
)

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect a client to a robot",
	Long: `Connect a client to the Erebus instance with a robot on the Erebus
instance.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("expected 2 arguments (client and robot)")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := getControlClient()
		rArgs := &pb.ControlMessage_ConnectClientToRobotRequest{
			ClientName: args[0],
			RobotName:  args[1],
		}
		res, err := client.ConnectClientToRobot(context.Background(), rArgs)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error connecting client")
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		switch res.Data.(type) {
		case *pb.ControlMessage_ConnectClientToRobotResponse_Error:
			fmt.Fprintln(os.Stderr, "Error connecting client")
			fmt.Fprintln(os.Stderr, res.GetError())
			os.Exit(1)
		case *pb.ControlMessage_ConnectClientToRobotResponse_Ok_:
		default:
			fmt.Fprintln(os.Stderr, "Error connecting client")
			fmt.Fprintln(os.Stderr, "Unexpected response from broker")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// connectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// connectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
