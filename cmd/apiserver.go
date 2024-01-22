package cmd

import (
	"github.com/oneblock-ai/steve/v2/pkg/debug"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/oneblock-ai/oneblock/pkg/server"
)

func NewAPIServer(ctx CommandContext) *cobra.Command {
	a := APIServerConfig{
		cmdCtx: ctx,
	}
	cmd := &cobra.Command{
		Use:   "api-server [flags]",
		Short: "Run an apis-server",
		RunE:  a.Run,
	}

	a.init(cmd)
	return cmd
}

type APIServerConfig struct {
	cmdCtx          CommandContext
	debugConfig     debug.Config
	Namespace       string
	Threadiness     int
	HTTPListenPort  int
	HTTPSListenPort int
}

func (a *APIServerConfig) Run(cmd *cobra.Command, _ []string) error {
	a.debugConfig.MustSetupDebug()

	ctx := cmd.Context()
	cfg := server.Options{
		Context:         ctx,
		KubeConfig:      a.cmdCtx.Kubeconfig,
		Name:            a.cmdCtx.Name,
		HTTPListenPort:  a.HTTPListenPort,
		HTTPSListenPort: a.HTTPSListenPort,
		Namespace:       a.Namespace,
	}
	ob, err := server.New(cfg)
	if err != nil {
		logrus.Errorf("failed to init new server, error: %v", err)
		return err
	}

	return ob.ListenAndServe(nil)
}

func (a *APIServerConfig) init(apiCmd *cobra.Command) {
	f := apiCmd.Flags()
	f.StringVar(&a.Namespace, "namespace", "oneblock-system", "default namespace to store system resources")
	f.IntVar(&a.Threadiness, "threadiness", 5, "controller threads")
	f.IntVar(&a.HTTPListenPort, "http_port", 8080, "HTTP listen port")
	f.IntVar(&a.HTTPSListenPort, "https_port", 8443, "HTTPS listen port")
}
