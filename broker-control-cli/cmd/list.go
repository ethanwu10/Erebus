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

// getCmd represents the get command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List objects (robots and clients)",
	Long: `List objects (robots and clients) that are currently connected to this
Erebus instance`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires an object type")
		}
		if len(args) > 1 {
			return errors.New("too many arguments received; expected 1")
		}
		if strings.HasPrefix(args[0], "robot") || strings.HasPrefix(args[0], "client") {
			return nil
		}

		return fmt.Errorf("invalid object type \"%s\"", args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := getControlClient()
		if strings.HasPrefix(args[0], "robot") {
			robots, err := client.GetRobots(context.Background(), &pb.Null{})
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error getting robots")
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			for _, robot := range robots.GetRobotNames() {
				fmt.Println(robot)
			}
		}
		if strings.HasPrefix(args[0], "client") {
			clients, err := client.GetClientControllers(context.Background(), &pb.Null{})
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error getting clients")
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			for _, client := range clients.GetControllerNames() {
				fmt.Println(client)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.
}
