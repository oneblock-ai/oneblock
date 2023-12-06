package config

import (
	"context"

	dashboardapi "github.com/kubernetes/dashboard/src/app/backend/auth/api"
	corev1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/core"
	rbacv1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/rbac"
	"github.com/rancher/wrangler/v2/pkg/start"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/oneblock-ai/oneblock/pkg/auth"
	obcorev1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/core.oneblock.ai"
	obmgmtv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/management.oneblock.ai"
)

type Management struct {
	ctx        context.Context
	Namespace  string
	ClientSet  *kubernetes.Clientset
	RestConfig *rest.Config

	OneBlockCoreFactory *obcorev1.Factory
	OneBlockMgmtFactory *obmgmtv1.Factory
	CoreFactory         *corev1.Factory
	RbacFactory         *rbacv1.Factory
	TokenManager        dashboardapi.TokenManager

	starters []start.Starter
}

func SetupManagement(ctx context.Context, restConfig *rest.Config, namespace string) (*Management, error) {
	mgmt := &Management{
		ctx:       ctx,
		Namespace: namespace,
	}

	mgmt.RestConfig = restConfig
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	mgmt.ClientSet = clientSet

	oneblockCore, err := obcorev1.NewFactoryFromConfig(restConfig)
	if err != nil {
		return nil, err
	}
	mgmt.OneBlockCoreFactory = oneblockCore
	mgmt.starters = append(mgmt.starters, oneblockCore)

	oneblockMgmt, err := obmgmtv1.NewFactoryFromConfig(restConfig)
	if err != nil {
		return nil, err
	}
	mgmt.OneBlockMgmtFactory = oneblockMgmt
	mgmt.starters = append(mgmt.starters, oneblockMgmt)

	core, err := corev1.NewFactoryFromConfig(restConfig)
	if err != nil {
		return nil, err
	}
	mgmt.CoreFactory = core
	mgmt.starters = append(mgmt.starters, core)

	rbac, err := rbacv1.NewFactoryFromConfig(restConfig)
	if err != nil {
		return nil, err
	}
	mgmt.RbacFactory = rbac
	mgmt.starters = append(mgmt.starters, rbac)

	mgmt.TokenManager, err = auth.NewJWETokenManager(core.Core().V1().Secret(), namespace)
	if err != nil {
		return nil, err
	}

	return mgmt, nil
}

func (m *Management) Start(threadiness int) error {
	return start.All(m.ctx, threadiness, m.starters...)
}
