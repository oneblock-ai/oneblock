package cmd

import (
	"github.com/oneblock-ai/steve/v2/pkg/debug"
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
	Kubeconfig             string
	Name                   string
	Version                string
	Namespace              string
	Threadiness            int
	HTTPListenPort         int
	HTTPSListenPort        int
	WebhookHTTPSListenPort int
	debugConfig            debug.Config

	cmdCtx CommandContext
}

func (a *APIServerConfig) Run(cmd *cobra.Command, _ []string) error {
	a.debugConfig.MustSetupDebug()

	ctx := cmd.Context()
	cfg := server.Options{
		Context:                ctx,
		KubeConfig:             a.cmdCtx.OneBlock.Kubeconfig,
		HTTPListenPort:         a.HTTPListenPort,
		HTTPSListenPort:        a.HTTPSListenPort,
		WebhookHTTPSListenPort: a.WebhookHTTPSListenPort,
		Namespace:              a.Namespace,
	}
	ob, err := server.New(cfg)
	if err != nil {
		return err
	}

	return ob.ListenAndServe(nil)
}

func (a *APIServerConfig) init(apiCmd *cobra.Command) {
	f := apiCmd.Flags()
	f.StringVar(&a.Name, "name", "oneblock-server", "name of the server")
	f.StringVar(&a.Version, "version", "dev", "version of the server")
	f.StringVar(&a.Namespace, "namespace", "oneblock-system", "default namespace to store system resources")
	f.IntVar(&a.Threadiness, "threadiness", 5, "controller threads")
	f.IntVar(&a.HTTPListenPort, "http_port", 8080, "HTTP listen port")
	f.IntVar(&a.HTTPSListenPort, "https_port", 8443, "HTTPS listen port")
	f.IntVar(&a.WebhookHTTPSListenPort, "webhook_https_port", 8444, "webhook HTTPS listen port")
}
