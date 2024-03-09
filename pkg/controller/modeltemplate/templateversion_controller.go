package modeltemplate

import (
	"context"
	"reflect"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mlv1 "github.com/oneblock-ai/oneblock/pkg/apis/ml.oneblock.ai/v1"
	ctloneblockv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/ml.oneblock.ai/v1"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
)

const (
	modelTmpVersionOnChange         = "modelTemplateVersion.onChange"
	modelTmpVersionConfigDeployment = "modelTemplateVersion.ConfigModelDeployment"
)

type TemplateVersionHandler struct {
	ctx                  context.Context
	releaseName          string
	templateVersion      ctloneblockv1.ModelTemplateVersionController
	templateVersionCache ctloneblockv1.ModelTemplateVersionCache
}

func VersionRegister(ctx context.Context, mgmt *config.Management) error {
	tmpVersion := mgmt.OneBlockMLFactory.Ml().V1().ModelTemplateVersion()
	handler := &TemplateVersionHandler{
		ctx:                  ctx,
		releaseName:          mgmt.ReleaseName,
		templateVersion:      tmpVersion,
		templateVersionCache: tmpVersion.Cache(),
	}

	tmpVersion.OnChange(ctx, modelTmpVersionOnChange, handler.OnChange)
	tmpVersion.OnChange(ctx, modelTmpVersionConfigDeployment, handler.ConfigModelDeployment)
	return nil
}

func (h *TemplateVersionHandler) OnChange(_ string, tv *mlv1.ModelTemplateVersion) (*mlv1.ModelTemplateVersion, error) {
	if tv == nil || tv.DeletionTimestamp != nil {
		return tv, nil
	}

	tvCpy := tv.DeepCopy()
	if tvCpy.Labels == nil {
		tvCpy.Labels = make(map[string]string)
	}
	tvCpy.Labels[constant.LabelModelTemplateName] = tv.Spec.TemplateName

	// add owner reference
	if tvCpy.OwnerReferences == nil {
		tvCpy.OwnerReferences = []metav1.OwnerReference{
			{
				Name:       tv.Name,
				APIVersion: tv.APIVersion,
				UID:        tv.UID,
				Kind:       tv.Kind,
			},
		}
	}

	if !reflect.DeepEqual(tvCpy, tv) {
		return h.templateVersion.Update(tvCpy)
	}
	return nil, nil
}

func (h *TemplateVersionHandler) ConfigModelDeployment(_ string, tv *mlv1.ModelTemplateVersion) (*mlv1.ModelTemplateVersion, error) {
	if tv == nil || tv.DeletionTimestamp != nil {
		return tv, nil
	}

	if mlv1.ModelTemplateVersionConfigured.IsTrue(tv) {
		logrus.Debugf("ModelTemplateVersion %s is already configured, skip updating", tv.Name)
		return nil, nil
	}

	tvCpy := tv.DeepCopy()
	modelConfig, err := generateRayLLMModelConfig(tv)
	if err != nil {
		mlv1.ModelTemplateVersionConfigured.SetError(tvCpy, "", err)
	} else {
		tvCpy.Status.GeneratedModelConfig = modelConfig
		mlv1.ModelTemplateVersionConfigured.SetError(tvCpy, "", nil)
	}
	if !reflect.DeepEqual(tvCpy.Status, tv.Status) {
		return h.templateVersion.UpdateStatus(tvCpy)
	}
	return nil, nil
}
