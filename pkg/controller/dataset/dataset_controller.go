package dataset

import (
	"context"
	"fmt"

	oneblockv1 "github.com/oneblock-ai/oneblock/pkg/apis/ml.oneblock.ai/v1"
	ctloneblockv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/ml.oneblock.ai/v1"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
)

const dsControllerName = "ob-dataset-controller"

type Handler struct {
	ctx      context.Context
	datasets ctloneblockv1.DatasetController
	dsCache  ctloneblockv1.DatasetCache
}

func Register(ctx context.Context, mgmt *config.Management) error {
	datasets := mgmt.OneBlockMLFactory.Ml().V1().Dataset()
	dsHandler := &Handler{
		ctx:      ctx,
		datasets: datasets,
		dsCache:  datasets.Cache(),
	}

	datasets.OnChange(ctx, dsControllerName, dsHandler.OnChange)
	datasets.OnRemove(ctx, dsControllerName, dsHandler.OnDelete)
	return nil
}

func (h *Handler) OnChange(_ string, dataset *oneblockv1.Dataset) (*oneblockv1.Dataset, error) {
	if dataset == nil || dataset.DeletionTimestamp != nil {
		return dataset, nil
	}
	fmt.Printf("dataset changed: %+v\n", dataset)

	return nil, nil
}

func (h *Handler) OnDelete(_ string, dataset *oneblockv1.Dataset) (*oneblockv1.Dataset, error) {
	if dataset == nil {
		return nil, nil
	}
	fmt.Printf("dataset on delete: %+v\n", dataset)
	return dataset, nil
}
