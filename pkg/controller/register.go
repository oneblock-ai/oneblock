package controller

import (
	"context"

	"github.com/rancher/wrangler/v2/pkg/leader"

	obAuth "github.com/oneblock-ai/oneblock/pkg/controller/auth"
	"github.com/oneblock-ai/oneblock/pkg/controller/dataset"
	"github.com/oneblock-ai/oneblock/pkg/controller/gpu"
	"github.com/oneblock-ai/oneblock/pkg/controller/mlservice"
	"github.com/oneblock-ai/oneblock/pkg/controller/modeltemplate"
	"github.com/oneblock-ai/oneblock/pkg/controller/notebook"
	"github.com/oneblock-ai/oneblock/pkg/controller/raycluster"
	"github.com/oneblock-ai/oneblock/pkg/controller/setting"
	"github.com/oneblock-ai/oneblock/pkg/controller/user"
	"github.com/oneblock-ai/oneblock/pkg/indexeres"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
)

const oneBlockRegisterControllersName = "oneblock-controllers"

type registerFunc func(context.Context, *config.Management) error

var registerFuncs = []registerFunc{
	indexeres.Register,
	setting.Register,
	dataset.Register,
	user.Register,
	raycluster.Register,
	gpu.Register,
	notebook.Register,
	modeltemplate.Register,
	mlservice.Register,
}

func register(ctx context.Context, mgmt *config.Management) error {
	for _, f := range registerFuncs {
		if err := f(ctx, mgmt); err != nil {
			return err
		}
	}

	obAuth.BootstrapAdminUser(mgmt)
	go obAuth.WatchSecret(ctx, mgmt)
	return nil
}

func Register(ctx context.Context, mgmt *config.Management, threadiness int) error {
	go leader.RunOrDie(ctx, "", oneBlockRegisterControllersName, mgmt.ClientSet, func(ctx context.Context) {
		if err := register(ctx, mgmt); err != nil {
			panic(err)
		}
		if err := mgmt.Start(threadiness); err != nil {
			panic(err)
		}
		<-ctx.Done()
	})
	return nil
}
