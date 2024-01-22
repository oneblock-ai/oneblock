package server

import (
	"context"

	wc "github.com/oneblock-ai/webhook/pkg/config"
	ws "github.com/oneblock-ai/webhook/pkg/server"
	"github.com/rancher/wrangler/v2/pkg/k8scheck"
	"github.com/rancher/wrangler/v2/pkg/ratelimit"
	"k8s.io/client-go/rest"

	sserver "github.com/oneblock-ai/oneblock/pkg/server"
	"github.com/oneblock-ai/oneblock/pkg/webhook"
)

const (
	webhookName = "oneblock-webhook"
)

// WebhookServer defines the webhook webhookServer types
type WebhookServer struct {
	ctx           context.Context
	webhookServer *ws.WebhookServer
	restConfig    *rest.Config
}

// Options define the api webhookServer options
type Options struct {
	Context         context.Context
	KubeConfig      string
	HTTPSListenPort int
	Threadiness     int
	Namespace       string
	Name            string
	DevMode         bool
	DevURL          string
}

func New(opts Options) (*WebhookServer, error) {
	s := &WebhookServer{
		ctx: opts.Context,
	}

	clientConfig, err := sserver.GetConfig(opts.KubeConfig)
	if err != nil {
		return s, err
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return s, err
	}
	restConfig.RateLimiter = ratelimit.None
	s.restConfig = restConfig

	err = k8scheck.Wait(s.ctx, *restConfig)
	if err != nil {
		return nil, err
	}

	// set up a new webhook webhookServer
	s.webhookServer = ws.NewWebhookServer(opts.Context, restConfig, webhookName, &wc.Options{
		Namespace:       opts.Namespace,
		Threadiness:     opts.Threadiness,
		HTTPSListenPort: opts.HTTPSListenPort,
		DevMode:         opts.DevMode,
		DevURL:          opts.DevURL,
	})

	if err := webhook.Register(opts.Context, restConfig, s.webhookServer, opts.Name, opts.Threadiness); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *WebhookServer) ListenAndServe() error {
	if err := s.webhookServer.Start(); err != nil {
		return err
	}

	<-s.ctx.Done()
	return nil
}
