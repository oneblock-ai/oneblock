package raycluster

import (
	"fmt"
	"strings"

	"github.com/oneblock-ai/webhook/pkg/server/admission"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"github.com/sirupsen/logrus"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ctlrayv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/ray.io/v1"
	"github.com/oneblock-ai/oneblock/pkg/webhook/config"
)

type validator struct {
	admission.DefaultValidator
	rayClusterCache ctlrayv1.RayClusterCache
}

var _ admission.Validator = &validator{}

func NewValidator(mgmt *config.Management) admission.Validator {
	return &validator{
		rayClusterCache: mgmt.KubeRayFactory.Ray().V1().RayCluster().Cache(),
	}
}

func (v *validator) Create(_ *admission.Request, newObj runtime.Object) error {
	cluster := newObj.(*rayv1.RayCluster)

	logrus.Debugf("[webhook validating]raycluster %s is created", cluster.Name)

	return checkRayVersionIsConsistent(cluster)
}

func (v *validator) Update(_ *admission.Request, _, newObj runtime.Object) error {
	cluster := newObj.(*rayv1.RayCluster)

	logrus.Debugf("[webhook validating]raycluster %s is updated", cluster.Name)

	return checkRayVersionIsConsistent(cluster)
}

func checkRayVersionIsConsistent(cluster *rayv1.RayCluster) error {
	rayVersion := cluster.Spec.RayVersion
	headContainer := cluster.Spec.HeadGroupSpec.Template.Spec.Containers[0]
	if err := validateImageVersion(headContainer.Image, rayVersion); err != nil {
		return err
	}

	workerSpecs := cluster.Spec.WorkerGroupSpecs
	for _, spec := range workerSpecs {
		for _, c := range spec.Template.Spec.Containers {
			if err := validateImageVersion(c.Image, rayVersion); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateImageVersion(image string, version string) error {
	if !strings.Contains(image, version) {
		return fmt.Errorf("image: %s is not consistent with cluster ray version %s", image, version)
	}
	return nil
}

func (v *validator) Resource() admission.Resource {
	return admission.Resource{
		Names:      []string{"rayclusters"},
		Scope:      admissionregv1.NamespacedScope,
		APIGroup:   rayv1.SchemeGroupVersion.Group,
		APIVersion: rayv1.SchemeGroupVersion.Version,
		ObjectType: &rayv1.RayCluster{},
		OperationTypes: []admissionregv1.OperationType{
			admissionregv1.Create,
			admissionregv1.Update,
		},
	}
}
