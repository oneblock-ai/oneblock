package queue

import (
	"net/http"

	"github.com/oneblock-ai/apiserver/v2/pkg/types"
	"github.com/oneblock-ai/steve/v2/pkg/schema"
	"github.com/oneblock-ai/steve/v2/pkg/server"
	"github.com/rancher/wrangler/v2/pkg/schemas"

	ctlschedulv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/scheduling.volcano.sh/v1beta1"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
)

const (
	queueSchemaID = "scheduling.volcano.sh.queue"

	ActionSetDefault = "setDefault"
)

type Handler struct {
	httpClient http.Client
	queue      ctlschedulv1.QueueClient
	queueCache ctlschedulv1.QueueCache
}

func RegisterSchema(mgmt *config.Management, server *server.Server) error {
	queues := mgmt.SchedulingFactory.Scheduling().V1beta1().Queue()
	h := Handler{
		httpClient: http.Client{},
		queue:      queues,
		queueCache: queues.Cache(),
	}

	t := []schema.Template{
		{
			ID:        queueSchemaID,
			Formatter: formatter,
			Customize: func(apiSchema *types.APISchema) {
				apiSchema.ResourceActions = map[string]schemas.Action{
					ActionSetDefault: {},
				}
				apiSchema.ActionHandlers = map[string]http.Handler{
					ActionSetDefault: h,
				}
			},
		},
	}

	server.SchemaFactory.AddTemplate(t...)
	return nil
}
