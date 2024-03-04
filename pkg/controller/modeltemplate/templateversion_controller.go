package modeltemplate

import (
	"context"
	"reflect"

	"github.com/sirupsen/logrus"

	mlv1 "github.com/oneblock-ai/oneblock/pkg/apis/ml.oneblock.ai/v1"
	ctloneblockv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/ml.oneblock.ai/v1"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
)

const (
	mlTemplateVersionControllerOnChange = "modelTemplateVersion.onChange"
)

type Handler struct {
	ctx                  context.Context
	releaseName          string
	templateVersion      ctloneblockv1.ModelTemplateVersionController
	templateVersionCache ctloneblockv1.ModelTemplateVersionCache
}

func Register(ctx context.Context, mgmt *config.Management) error {
	tmpVersion := mgmt.OneBlockMLFactory.Ml().V1().ModelTemplateVersion()
	handler := &Handler{
		ctx:                  ctx,
		releaseName:          mgmt.ReleaseName,
		templateVersion:      tmpVersion,
		templateVersionCache: tmpVersion.Cache(),
	}

	tmpVersion.OnChange(ctx, mlTemplateVersionControllerOnChange, handler.OnChange)
	return nil
}

func (h *Handler) OnChange(_ string, templateVersion *mlv1.ModelTemplateVersion) (*mlv1.ModelTemplateVersion, error) {
	if templateVersion == nil || templateVersion.DeletionTimestamp != nil {
		return templateVersion, nil
	}

	if mlv1.ModelTemplateVersionConfigured.IsTrue(templateVersion) {
		logrus.Debugf("ModelTemplateVersion %s is already configured, skip updating", templateVersion.Name)
		return nil, nil
	}

	tmpVersionCpy := templateVersion.DeepCopy()
	modelConfig, err := generateRayLLMModelConfig(templateVersion)
	if err != nil {
		mlv1.ModelTemplateVersionConfigured.False(tmpVersionCpy)
		mlv1.ModelTemplateVersionConfigured.Message(tmpVersionCpy, err.Error())
	} else {
		tmpVersionCpy.Status.GeneratedModelConfig = modelConfig
		mlv1.ModelTemplateVersionConfigured.True(tmpVersionCpy)
		mlv1.ModelTemplateVersionConfigured.Message(tmpVersionCpy, "")
	}
	if !reflect.DeepEqual(tmpVersionCpy.Status, templateVersion.Status) {
		return h.templateVersion.UpdateStatus(tmpVersionCpy)
	}
	return templateVersion, err
}
