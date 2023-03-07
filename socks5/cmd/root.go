package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "socks5-server",
	Short: "start a socks5 server",
}

func Execute() error {
	return rootCmd.Execute()
}

var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(startCmd)
}
