package main

import (
	cmd "github.com/mosaicnetworks/evm-babble/cmd/evm-babble/commands"
)

func main() {

	rootCmd := cmd.RootCmd

	rootCmd.AddCommand(
		cmd.NewRunCmd(),
		cmd.VersionCmd)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
