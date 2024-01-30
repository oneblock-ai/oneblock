package cmd

import (
	"fmt"

	"github.com/oneblock-ai/steve/v2/pkg/debug"
	"github.com/spf13/cobra"

	wserver "github.com/oneblock-ai/oneblock/pkg/webhook/server"
)

func NewWebhookServer(ctx CommandContext) *cobra.Command {
	a := WebhookConfig{
		cmdCtx: ctx,
	}
	cmd := &cobra.Command{
		Use:   "webhook [flags]",
		Short: "Run an webhook server",
		RunE:  a.Run,
	}

	a.init(cmd)
	return cmd
}

type WebhookConfig struct {
	cmdCtx          CommandContext
	debugConfig     debug.Config
	Namespace       string
	Threadiness     int
	HTTPSListenPort int
	DevMode         bool
	DevURL          string
}

func (a *WebhookConfig) Run(cmd *cobra.Command, _ []string) error {
	a.debugConfig.MustSetupDebug()

	ctx := cmd.Context()
	cfg := wserver.Options{
		Context:         ctx,
		KubeConfig:      a.cmdCtx.Kubeconfig,
		Name:            a.cmdCtx.Name,
		HTTPSListenPort: a.HTTPSListenPort,
		Namespace:       a.Namespace,
		Threadiness:     a.Threadiness,
		DevMode:         a.DevMode,
		DevURL:          a.DevURL,
	}
	ws, err := wserver.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to init webhook server, error: %v", err)
	}

	return ws.ListenAndServe()
}

func (a *WebhookConfig) init(apiCmd *cobra.Command) {
	f := apiCmd.Flags()
	f.IntVar(&a.HTTPSListenPort, "https_port", 8444, "webhook HTTPS listen port")
	f.StringVar(&a.Namespace, "namespace", "oneblock-system", "default namespace to store system resources")
	f.IntVar(&a.Threadiness, "threadiness", 5, "controller threads")
	f.BoolVar(&a.DevMode, "dev_mode", false, "enable local dev mode")
	f.StringVar(&a.DevURL, "dev_url", "", "specify the webhook local url, only used when dev_mode is true")
}
