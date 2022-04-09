package main

import (
	"github.com/spf13/cobra"
)

func main() {
	Execute()
}

type ConfigArgs struct {
	HostsArg string // required
	PortsArg string // required

	Shuffle            bool   // default false, shuffle the queue of scan
	TimeoutInSecondArg int    // default 3s
	ThreadArg          int    // default 1
	OutputArg          string // default none
}

var configArgs ConfigArgs

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "port-scanner",
	Short: "A fast port scan tool based on full tcp connection",
	// Run: func(cmd *cobra.Command, args []string) { }, // todo implement
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// todo add flags
}
