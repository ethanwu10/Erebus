package cmd

import (
	"github.com/spf13/cobra"
)

// simCmd represents the sim command
var simCmd = &cobra.Command{
	Use:   "sim",
	Short: "Control the sim state",
	Long:  `Control the state of the simulation in the running Erebus instance`,
}

func init() {
	rootCmd.AddCommand(simCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// simCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// simCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
