package user

import (
	"github.com/oneblock-ai/webhook/pkg/server/admission"
	"github.com/sirupsen/logrus"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"

	managementv1 "github.com/oneblock-ai/oneblock/pkg/apis/management.oneblock.ai/v1"
)

type mutator struct {
	admission.DefaultMutator
}

var _ admission.Mutator = &mutator{}

func NewMutator() admission.Mutator {
	return &mutator{}
}

func (m *mutator) Create(_ *admission.Request, newObj runtime.Object) (admission.Patch, error) {
	user := newObj.(*managementv1.User)
	logrus.Infof("[webhook mutating]user %s is created", user.Name)

	labels := user.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	labels["oneblock.ai/creator"] = "oneblock"

	return admission.Patch{admission.PatchOp{
		Op:    admission.PatchOpReplace,
		Path:  "/metadata/labels",
		Value: labels,
	}}, nil
}

func (m *mutator) Resource() admission.Resource {
	return admission.Resource{
		Names:      []string{"users"},
		Scope:      admissionregv1.ClusterScope,
		APIGroup:   managementv1.SchemeGroupVersion.Group,
		APIVersion: managementv1.SchemeGroupVersion.Version,
		ObjectType: &managementv1.User{},
		OperationTypes: []admissionregv1.OperationType{
			admissionregv1.Create,
		},
	}
}
