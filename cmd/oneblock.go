package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	o := &OneBlockOptions{}
	var rootCmd = &cobra.Command{
		Use:  "oneblock",
		Long: "An open-source cloud-native LLMOps(Large Language Model Operations) platform.",
		RunE: Run,
	}
	cmdContext := CommandContext{
		OneBlock: o,
		StdOut:   os.Stdout,
		StdErr:   os.Stderr,
		StdIn:    nil,
	}

	// initialize the root command configs
	o.init(rootCmd)

	rootCmd.AddCommand(
		NewAPIServer(cmdContext),
		NewVersion(cmdContext),
	)

	rootCmd.InitDefaultHelpCmd()
	return rootCmd
}

// OneBlockOptions define the common struct for all commands
type OneBlockOptions struct {
	Kubeconfig string `usage:"kubeconfig file"`
	Project    string `usage:"project name" short:"p"`
	Debug      bool   `usage:"enable debug mode" short:"d"`
	DebugLevel int    `usage:"debug level" short:"v" default:"7"`
	Trace      bool   `usage:"enable trace mode" short:"t"`
}

func (o *OneBlockOptions) init(cmd *cobra.Command) {
	f := cmd.PersistentFlags()
	f.StringVar(&o.Kubeconfig, "kubeconfig", "", "kubeconfig file path")
	f.StringVarP(&o.Project, "project", "p", "", "project name")
	f.BoolVarP(&o.Debug, "debug", "d", false, "enable debug mode")
	f.IntVar(&o.DebugLevel, "debug_level", 7, "debug level")
	f.BoolVarP(&o.Trace, "trace", "t", false, "enable trace mode")
}

func Run(_ *cobra.Command, _ []string) error {
	fmt.Println("run oneblock command")
	return nil
}
