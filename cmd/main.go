package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func main() {
	Execute()
}

type ConfigArgs struct {
	HostsArg string // required
	PortsArg string // required

	Shuffle            bool   // default false, shuffle the queue of scan
	TimeoutInSecondArg int    // default 3 s
	ThreadArg          int    // default 20
	OutputArg          string // default none

	PrintVersion bool
}

var configArgs ConfigArgs

var name = filepath.Base(os.Args[0])
var version = "v0.0.0"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:  "port-scanner",
	Long: "A fast port scan tool based on full tcp connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		if configArgs.PrintVersion {
			fmt.Println(version)
			return nil
		}

		if configArgs.HostsArg == "" || configArgs.PortsArg == "" {
			return errors.New(`required flag(s) "hosts" or "ports" not set`)
		}

		newHandler(configArgs).handle()

		return nil
	},
	Example: fmt.Sprintf(`  1. %s -h 192.168.*.0 -p 80,443
  2. %s -h 192.168.*.0 -p 80,443 -t 1 -n 50 -s -o result.txt`, name, name),
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().BoolP("help", "", false, "help for this command")
	rootCmd.Flags().StringVarP(&configArgs.HostsArg,
		"hosts", "h", "", "hosts, eg: 10.0.0.1,10.0.0.5-10,192.168.1.*,192.168.10.0/24")
	rootCmd.Flags().StringVarP(&configArgs.PortsArg,
		"ports", "p", "", "ports, eg: 80 or 1-1024 or 1-1024,3389,8080")
	rootCmd.Flags().BoolVarP(&configArgs.Shuffle,
		"shuffle", "s", false, "shuffle hosts and ports")
	rootCmd.Flags().IntVarP(&configArgs.TimeoutInSecondArg,
		"timeout", "t", 3, "timeout(second) of tcp connect")
	rootCmd.Flags().IntVarP(&configArgs.ThreadArg,
		"thread", "n", 20, "thread number")
	rootCmd.Flags().StringVarP(&configArgs.OutputArg,
		"output", "o", "", "file path to output the opened ports")
	rootCmd.Flags().BoolVarP(&configArgs.PrintVersion,
		"version", "v", false, "version")
}
