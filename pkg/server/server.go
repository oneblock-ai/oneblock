package server

import (
	"context"

	dlserver "github.com/oneblock-ai/dynamiclistener/v2/server"
	"github.com/oneblock-ai/steve/v2/pkg/server"
	"github.com/rancher/wrangler/v2/pkg/ratelimit"
	"k8s.io/client-go/rest"

	oneblockAuth "github.com/oneblock-ai/oneblock/pkg/api/auth"
	"github.com/oneblock-ai/oneblock/pkg/controller"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
)

// Server defines the api server types
type Server struct {
	ctx             context.Context
	kubeconfig      string
	httpListenPort  int
	httpsListenPort int
	threadiness     int
	namespace       string

	mgmt        *config.Management
	steveServer *server.Server
	restConfig  *rest.Config
}

// Options define the api server options
type Options struct {
	Context         context.Context
	KubeConfig      string
	HTTPListenPort  int
	HTTPSListenPort int
	Threadiness     int
	Namespace       string
}

func (s *Server) setDefaults(cfg *rest.Config) (*server.Options, error) {
	var err error
	opts := &server.Options{}

	// set up the management config
	s.mgmt, err = config.SetupManagement(s.ctx, cfg, s.namespace)
	if err != nil {
		return nil, err
	}

	// define the next handler after mgmt
	r := NewRouter(s.mgmt)
	opts.Next = r.Routes()

	// define the custom auth middleware
	middleware := oneblockAuth.NewMiddleware(s.mgmt)
	opts.AuthMiddleware = middleware.AuthMiddleware

	return opts, nil
}

func New(opts Options) (*Server, error) {
	var err error
	s := &Server{
		ctx:             opts.Context,
		kubeconfig:      opts.KubeConfig,
		httpListenPort:  opts.HTTPListenPort,
		httpsListenPort: opts.HTTPSListenPort,
		namespace:       opts.Namespace,
		threadiness:     opts.Threadiness,
	}

	clientConfig, err := GetConfig(s.kubeconfig)
	if err != nil {
		return nil, err
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	restConfig.RateLimiter = ratelimit.None
	s.restConfig = restConfig

	serverOptions, err := s.setDefaults(restConfig)
	if err != nil {
		return nil, err
	}

	// set up a new steve server
	s.steveServer, err = server.New(opts.Context, restConfig, serverOptions)
	if err != nil {
		return nil, err
	}

	if err := controller.Register(s.ctx, s.mgmt, s.threadiness); err != nil {
		return s, err
	}

	return s, nil
}

func (s *Server) ListenAndServe(opts *dlserver.ListenOpts) error {
	return s.steveServer.ListenAndServe(s.ctx, s.httpsListenPort, s.httpListenPort, opts)
}
