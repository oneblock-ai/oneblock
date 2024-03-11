package api

import (
	"context"

	"github.com/oneblock-ai/steve/v2/pkg/server"

	"github.com/oneblock-ai/oneblock/pkg/api/queue"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
)

type registerSchema func(mgmt *config.Management, server *server.Server) error

func registerSchemas(mgmt *config.Management, server *server.Server, registers ...registerSchema) error {
	for _, register := range registers {
		if err := register(mgmt, server); err != nil {
			return err
		}
	}
	return nil
}

func Register(_ context.Context, mgmt *config.Management, server *server.Server) error {
	return registerSchemas(mgmt, server,
		queue.RegisterSchema)
}
