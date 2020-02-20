package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	//homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

var server string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "erebus-broker-control",
	Short: "CLI for interacting with the Erebus broker",
	Long: `This program sends commands to control the Erebus broker, which allows
for the control of the Erebus simulation environment and the management
of competitors and robots connected to this Erebus instance.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.broker-control-cli.yaml)")

	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "localhost:51512", "Erebus server to connect to")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	//	if cfgFile != "" {
	//		// Use config file from the flag.
	//		viper.SetConfigFile(cfgFile)
	//	} else {
	//		// Find home directory.
	//		home, err := homedir.Dir()
	//		if err != nil {
	//			fmt.Println(err)
	//			os.Exit(1)
	//		}
	//
	//		// Search config in home directory with name ".broker-control-cli" (without extension).
	//		viper.AddConfigPath(home)
	//		viper.SetConfigName(".broker-control-cli")
	//	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
