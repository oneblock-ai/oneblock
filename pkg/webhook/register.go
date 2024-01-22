package webhook

import (
	"context"
	"fmt"

	ws "github.com/oneblock-ai/webhook/pkg/server"
	"github.com/oneblock-ai/webhook/pkg/server/admission"
	"k8s.io/client-go/rest"

	wconfig "github.com/oneblock-ai/oneblock/pkg/webhook/config"
	rayCluster "github.com/oneblock-ai/oneblock/pkg/webhook/resources/kuberay/raycluster"
	"github.com/oneblock-ai/oneblock/pkg/webhook/resources/user"
)

func register(mgmt *wconfig.Management) (validators []admission.Validator, mutators []admission.Mutator) {
	validators = []admission.Validator{
		user.NewValidator(mgmt),
		rayCluster.NewValidator(mgmt),
	}

	mutators = []admission.Mutator{
		user.NewMutator(),
		rayCluster.NewMutator(mgmt),
	}

	return
}

func Register(ctx context.Context, restConfig *rest.Config, ws *ws.WebhookServer, name string, threadiness int) error {
	// Separated factories are needed for the webhook register.
	// Controllers are running in active/standby mode. If the webhook register and controllers are use the same factories,
	// when the standby pod is upgraded to be active, it will be unable to add handlers and indexers to the controllers
	// because the factories are already started.
	mgmt, err := wconfig.SetupManagement(ctx, restConfig, name)
	if err != nil {
		return fmt.Errorf("setup management failed: %w", err)
	}

	validators, mutators := register(mgmt)

	if err := ws.RegisterValidators(validators...); err != nil {
		return fmt.Errorf("register validators failed: %w", err)
	}

	if err := ws.RegisterMutators(mutators...); err != nil {
		return fmt.Errorf("register mutators failed: %w", err)
	}

	if err := mgmt.Start(threadiness); err != nil {
		return fmt.Errorf("start management failed: %w", err)
	}

	return nil
}
