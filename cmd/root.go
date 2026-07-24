package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/calgorr/sms-gateway/config"
)

// rootCMD represents the base command when called without any subcommands
var rootCMD = &cobra.Command{
	Use:   "sms-gateway",
	Short: "SMS Gateway is the service for sending SMS messages",
}

// configPath is the path to the config file
var configPath string

func init() {
	rootCMD.PersistentFlags().StringVar(&configPath, "config", "", "config file")

	config.Init(configPath)

	rootCMD.AddCommand(serveApiCMD)
	rootCMD.AddCommand(workerExpressCMD)
	rootCMD.AddCommand(workerNormalCMD)
}

func Execute() {
	if err := rootCMD.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
