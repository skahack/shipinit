package main

import (
	"fmt"
	"os"

	"github.com/SKAhack/shipinit/cmd"
	"github.com/spf13/cobra"
)

var (
	Version  string
	Revision string
)

const (
	cliName        = "shipinit"
	cliDescription = "container initialization commands"
)

var rootCmd = &cobra.Command{
	Use:   cliName,
	Short: cliDescription,
}

func main() {
	rootCmd.AddCommand(
		cmd.NewEnvloadCommand(os.Stdout, os.Stderr),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
