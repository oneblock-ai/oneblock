package modeltemplate

import (
	"fmt"

	"github.com/oneblock-ai/webhook/pkg/server/admission"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"

	mlv1 "github.com/oneblock-ai/oneblock/pkg/apis/ml.oneblock.ai/v1"
)

type validator struct {
	admission.DefaultValidator
}

var _ admission.Validator = &validator{}

func NewValidator() admission.Validator {
	return &validator{}
}

func (v *validator) Create(_ *admission.Request, newObj runtime.Object) error {
	modelTemplateVersion := newObj.(*mlv1.ModelTemplateVersion)

	return validateModelPathConfig(modelTemplateVersion)
}

func (v *validator) Update(_ *admission.Request, _, newObj runtime.Object) error {
	modelTemplateVersion := newObj.(*mlv1.ModelTemplateVersion)

	return validateModelPathConfig(modelTemplateVersion)
}

func validateModelPathConfig(modelTmpVersion *mlv1.ModelTemplateVersion) error {
	if modelTmpVersion.Spec.HFModelID != "" && modelTmpVersion.Spec.MirrorConfig != "" {
		return fmt.Errorf("can't set both HF model ID or mirror config at the same time")
	}
	return nil
}

func (v *validator) Resource() admission.Resource {
	return admission.Resource{
		Names:      []string{"modelTemplateVersions"},
		Scope:      admissionregv1.NamespacedScope,
		APIGroup:   mlv1.SchemeGroupVersion.Group,
		APIVersion: mlv1.SchemeGroupVersion.Version,
		ObjectType: &mlv1.ModelTemplateVersion{},
		OperationTypes: []admissionregv1.OperationType{
			admissionregv1.Create,
			admissionregv1.Update,
		},
	}
}
