package config

import (
	"context"

	"github.com/rancher/wrangler/v2/pkg/start"
	"k8s.io/client-go/rest"

	obmgmtv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/management.oneblock.ai"
)

type Management struct {
	ctx        context.Context
	RestConfig *rest.Config

	OneBlockMgmtFactory *obmgmtv1.Factory
	starters            []start.Starter
}

func SetupManagement(ctx context.Context, restConfig *rest.Config) (*Management, error) {
	mgmt := &Management{
		ctx:        ctx,
		RestConfig: restConfig,
	}

	var err error
	mgmt.OneBlockMgmtFactory, err = obmgmtv1.NewFactoryFromConfig(restConfig)
	if err != nil {
		return nil, err
	}
	mgmt.starters = append(mgmt.starters, mgmt.OneBlockMgmtFactory)

	return mgmt, nil
}

func (m *Management) Start(threadiness int) error {
	return start.All(m.ctx, threadiness, m.starters...)
}
