package server

import (
	"context"

	"github.com/oneblock-ai/steve/v2/pkg/server"
	dlserver "github.com/rancher/dynamiclistener/server"
	"github.com/rancher/wrangler/v2/pkg/k8scheck"
	"github.com/rancher/wrangler/v2/pkg/ratelimit"
	"k8s.io/client-go/rest"

	apischema "github.com/oneblock-ai/oneblock/pkg/api"
	oneblockAuth "github.com/oneblock-ai/oneblock/pkg/api/auth"
	"github.com/oneblock-ai/oneblock/pkg/controller"
	"github.com/oneblock-ai/oneblock/pkg/data"
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
	releaseName     string

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
	Name            string
}

func (s *Server) setDefaults(cfg *rest.Config) (*server.Options, error) {
	var err error
	opts := &server.Options{}

	// set up the management config
	s.mgmt, err = config.SetupManagement(s.ctx, cfg, s.namespace, s.releaseName)
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
		releaseName:     opts.Name,
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

	err = k8scheck.Wait(s.ctx, *restConfig)
	if err != nil {
		return nil, err
	}

	serverOptions, err := s.setDefaults(restConfig)
	if err != nil {
		return nil, err
	}

	// set up a new steve server
	s.steveServer, err = server.New(opts.Context, restConfig, serverOptions)
	if err != nil {
		return nil, err
	}

	if err = controller.Register(opts.Context, s.mgmt, opts.Threadiness); err != nil {
		return nil, err
	}

	if err = apischema.Register(opts.Context, s.mgmt, s.steveServer); err != nil {
		return nil, err
	}

	return s, data.Init(s.ctx, s.mgmt, opts.Name)
}

func (s *Server) ListenAndServe(opts *dlserver.ListenOpts) error {
	return s.steveServer.ListenAndServe(s.ctx, s.httpsListenPort, s.httpListenPort, opts)
}
