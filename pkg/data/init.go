package data

import (
	"context"

	"github.com/oneblock-ai/oneblock/pkg/server/config"
)

// Init adds built-in resources
func Init(ctx context.Context, mgmt *config.Management, name string) error {
	if err := addPublicNamespace(mgmt.Apply); err != nil {
		return err
	}

	return addDefaultPublicRayCluster(ctx, mgmt, name)
}
