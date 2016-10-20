package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/graysonchao/netmux/netmux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "netmux",
	Short: "Multiplex a UDP stream across one or several sources.",
	Long: `netmux multiplexes a UDP stream across several sources.
These can be named pipes, UDP listeners, or Unix domain sockets.

The destinations to broadcast to are defined in the config file.

Usage:
netmux --config <path/to/config.json>`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.New(os.Stdout, "[netmux] ", log.Lshortfile|log.LstdFlags)
		logger.Printf("Using config file: %s", cfgFile)

		netmux.Start(logger)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.netmux.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName(".netmux") // name of config file (without extension)
	viper.AddConfigPath("$HOME")   // adding home directory as first search path
	viper.AutomaticEnv()           // read in environment variables that match

	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
	}
}
