package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	pb "github.com/ethanwu10/erebus/broker-control-cli/gen"
	"github.com/spf13/cobra"
)

// disconnectCmd represents the disconnect command
var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect a client from a robot",
	Long: `Disconnect a client to the Erebus instance from a robot on the Erebus
instance.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("expected 1 argument (client name)")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := getControlClient()
		rArgs := &pb.ControlMessage_DisconnectClientFromRobotRequest{
			ClientName: args[0],
		}
		res, err := client.DisconnectClientFromRobot(context.Background(), rArgs)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error disconnecting client")
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		switch res.Data.(type) {
		case *pb.ControlMessage_DisconnectClientFromRobotResponse_Error:
			fmt.Fprintln(os.Stderr, "Error disconnecting client")
			fmt.Fprintln(os.Stderr, res.GetError())
			os.Exit(1)
		case *pb.ControlMessage_DisconnectClientFromRobotResponse_Ok_:
		default:
			fmt.Fprintln(os.Stderr, "Error connecting client")
			fmt.Fprintln(os.Stderr, "Unexpected response from broker")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(disconnectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// disconnectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// disconnectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
