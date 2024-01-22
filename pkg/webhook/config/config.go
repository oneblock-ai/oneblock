package config

import (
	"context"

	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/wrangler/v2/pkg/generic"
	"github.com/rancher/wrangler/v2/pkg/start"
	"k8s.io/client-go/rest"

	obmgmtv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/management.oneblock.ai"
	kuberayv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/ray.io"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
)

type Management struct {
	ctx         context.Context
	ReleaseName string
	RestConfig  *rest.Config

	OneBlockMgmtFactory *obmgmtv1.Factory
	KubeRayFactory      *kuberayv1.Factory
	starters            []start.Starter
}

func SetupManagement(ctx context.Context, restConfig *rest.Config, releaseName string) (*Management, error) {
	mgmt := &Management{
		ctx:         ctx,
		RestConfig:  restConfig,
		ReleaseName: releaseName,
	}

	factory, err := controller.NewSharedControllerFactoryFromConfig(mgmt.RestConfig, config.Scheme)
	if err != nil {
		return nil, err
	}

	factoryOpts := &generic.FactoryOptions{
		SharedControllerFactory: factory,
	}

	mgmt.OneBlockMgmtFactory, err = obmgmtv1.NewFactoryFromConfig(restConfig)
	if err != nil {
		return nil, err
	}
	mgmt.starters = append(mgmt.starters, mgmt.OneBlockMgmtFactory)

	kuberay, err := kuberayv1.NewFactoryFromConfigWithOptions(restConfig, factoryOpts)
	if err != nil {
		return nil, err
	}
	mgmt.KubeRayFactory = kuberay
	mgmt.starters = append(mgmt.starters, kuberay)

	return mgmt, nil
}

func (m *Management) Start(threadiness int) error {
	return start.All(m.ctx, threadiness, m.starters...)
}
