package cmd

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // import pprof
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	o := &Oneblock{}
	rootCmd := &cobra.Command{
		Use:  "oneblock",
		Long: "An open-source cloud-native LLMOps(Large Language Model Operations) platform.",
		RunE: Run,
	}
	cmdContext := CommandContext{
		Oneblock: o,
		StdOut:   os.Stdout,
		StdErr:   os.Stderr,
		StdIn:    nil,
	}

	rootCmd.AddCommand(
		NewAPIServer(cmdContext),
		NewWebhookServer(cmdContext),
		NewVersion(),
	)
	// initialize the root command configs
	o.init(rootCmd)

	rootCmd.InitDefaultHelpCmd()
	return rootCmd
}

// Oneblock define the common struct for all commands
type Oneblock struct {
	Kubeconfig     string
	Name           string
	Version        string
	ProfileAddress string
	LogFormat      string
	Debug          bool
	Trace          bool
}

func (o *Oneblock) init(cmd *cobra.Command) {
	f := cmd.PersistentFlags()
	f.StringVar(&o.Kubeconfig, "kubeconfig", "", "kubeconfig file path")
	f.StringVar(&o.Name, "name", "oneblock", "oneblock release name")
	f.StringVar(&o.Version, "version", "dev", "version of the server")
	f.StringVar(&o.ProfileAddress, "profile_address", "0.0.0.0:6060", "address to listen on for profiling")
	f.StringVar(&o.LogFormat, "log_format", "text", "config log format")
	f.BoolVar(&o.Debug, "debug", false, "enable debug logs")
	f.BoolVar(&o.Trace, "trace", false, "enable trace logs")

	initProfiling(o)
	initLogs(o)
}

func Run(_ *cobra.Command, _ []string) error {
	fmt.Println("Oneblock root command")
	return nil
}

func initProfiling(o *Oneblock) {
	// enable profiler
	if o.ProfileAddress != "" {
		go func() {
			server := http.Server{
				Addr: o.ProfileAddress,
				// fix G114: Use of net/http serve function that has no support for setting timeouts (gosec)
				// refer to https://app.deepsource.com/directory/analyzers/go/issues/GO-S2114
				ReadHeaderTimeout: 10 * time.Second,
			}
			log.Println(server.ListenAndServe())
		}()
	}
}

func initLogs(o *Oneblock) {
	switch o.LogFormat {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
	logrus.SetOutput(os.Stdout)
	if o.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debugf("Loglevel set to [%v]", logrus.DebugLevel)
	}
	if o.Trace {
		logrus.SetLevel(logrus.TraceLevel)
		logrus.Tracef("Loglevel set to [%v]", logrus.TraceLevel)
	}
}
