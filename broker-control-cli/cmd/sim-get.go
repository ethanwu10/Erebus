package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pb "github.com/ethanwu10/erebus/broker-control-cli/gen"
)

// simGetCmd represents the get command
var simGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the current sim state",
	Long:  `Get the current state of the simulation in the running Erebus instance`,
	Run: func(cmd *cobra.Command, args []string) {
		client := getControlClient()
		res, err := client.GetSimulationState(context.Background(), &pb.Null{})
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error getting simulation state")
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		switch res.GetState() {
		case pb.SimState_RESET:
			fmt.Println("stopped")
		case pb.SimState_START:
			fmt.Println("started")
		case pb.SimState_STOP:
			fmt.Println("paused")
		case pb.SimState_UNKOWN:
			fallthrough
		default:
			fmt.Println("[unknown]")
		}
	},
}

func init() {
	simCmd.AddCommand(simGetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
