package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	pb "github.com/ethanwu10/erebus/broker-control-cli/gen"
)

// simSetCmd represents the set command
var simSetCmd = &cobra.Command{
	Use:   "set (start|stop|reset)",
	Short: "Set the current sim state",
	Long:  `Set the current state of the simulation in the running Erebus instance`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("expected one argument")
		}
		switch strings.ToLower(args[0]) {
		case "start":
		case "stop":
		case "pause":
		case "reset":
		default:
			return fmt.Errorf("invalid state \"%s\"", args[0])
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := getControlClient()
		var newState *pb.SimState
		switch strings.ToLower(args[0]) {
		case "start":
			newState = &pb.SimState{State: pb.SimState_START}
		case "pause":
			fallthrough
		case "stop":
			newState = &pb.SimState{State: pb.SimState_STOP}
		case "reset":
			newState = &pb.SimState{State: pb.SimState_RESET}
		}
		_, err := client.SetSimulationState(context.Background(), newState)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error setting simulation state")
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	simCmd.AddCommand(simSetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
