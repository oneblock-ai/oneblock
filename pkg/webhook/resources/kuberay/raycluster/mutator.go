package raycluster

import (
	"fmt"

	"github.com/oneblock-ai/webhook/pkg/server/admission"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"github.com/sirupsen/logrus"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"

	clusterctl "github.com/oneblock-ai/oneblock/pkg/controller/kuberay/cluster"
	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
	"github.com/oneblock-ai/oneblock/pkg/webhook/config"
)

type mutator struct {
	admission.DefaultMutator
	releaseName string
}

var _ admission.Mutator = &mutator{}

func NewMutator(mgmt *config.Management) admission.Mutator {
	return &mutator{
		releaseName: mgmt.ReleaseName,
	}
}

func (m *mutator) Create(_ *admission.Request, newObj runtime.Object) (admission.Patch, error) {
	cluster := newObj.(*rayv1.RayCluster)
	logrus.Debugf("[webhook mutating]raycluster %s is created", cluster.Name)

	patchOps := make([]admission.PatchOp, 0)
	var err error
	if val, ok := cluster.Annotations[constant.AnnotationRayClusterEnableGCS]; ok && val == "true" {
		patchOps, err = patchClusterGCSConfig(m.releaseName, cluster)
		if err != nil {
			return nil, err
		}
	}

	return patchOps, nil
}

func (m *mutator) Update(_ *admission.Request, oldObj runtime.Object, newObj runtime.Object) (admission.Patch, error) {
	oldCluster := oldObj.(*rayv1.RayCluster)
	cluster := newObj.(*rayv1.RayCluster)
	logrus.Debugf("[webhook mutating]raycluster %s is created", cluster.Name)

	val, ok := cluster.Annotations[constant.AnnotationRayClusterEnableGCS]
	// turn off GCS config is not allowed
	if valOld, okOld := oldCluster.Annotations[constant.AnnotationRayClusterEnableGCS]; okOld && valOld == "false" {
		if ok && val == "true" {
			return nil, fmt.Errorf("cannot disable GCS config once it is enabled")
		}
	}

	patchOps := make([]admission.PatchOp, 0)
	if ok && val == "true" {
		patches, err := patchClusterGCSConfig(m.releaseName, cluster)
		if err != nil {
			return nil, err
		}
		patchOps = append(patchOps, patches...)
	}

	return patchOps, nil

}

func patchClusterGCSConfig(releaseName string, cluster *rayv1.RayCluster) ([]admission.PatchOp, error) {
	patches := make([]admission.PatchOp, 0)

	AnnoPatch := getRayFTAnnotationPatch(cluster)
	patches = append(patches, AnnoPatch)

	envPatch := getHeadGroupEnvPath(cluster, releaseName)
	patches = append(patches, envPatch)

	paramsPatch := getHeadGroupStartParamsPath(cluster)
	patches = append(patches, paramsPatch)

	return patches, nil
}

func getRayFTAnnotationPatch(cluster *rayv1.RayCluster) admission.PatchOp {
	//patch := admission.PatchOp{}
	annotations := cluster.Annotations
	annotations[constant.AnnotationRayFTEnabledKey] = "true"

	return admission.PatchOp{
		Op:    admission.PatchOpReplace,
		Path:  "/metadata/annotations",
		Value: annotations,
	}
}

func getHeadGroupEnvPath(cluster *rayv1.RayCluster, releaseName string) admission.PatchOp {
	headGroupEnv := cluster.Spec.HeadGroupSpec.Template.Spec.Containers[0].Env
	redisEnvConfig := clusterctl.GetHeadNodeRedisEnvConfig(releaseName, cluster.Namespace)
	if headGroupEnv == nil || len(headGroupEnv) == 0 {
		headGroupEnv = redisEnvConfig
	} else {
		headGroupEnv = append(headGroupEnv, redisEnvConfig...)
	}

	return admission.PatchOp{
		Op:    admission.PatchOpReplace,
		Path:  "/spec/headGroupSpec/template/spec/containers/0/env",
		Value: headGroupEnv,
	}
}

func getHeadGroupStartParamsPath(cluster *rayv1.RayCluster) admission.PatchOp {
	startParams := cluster.Spec.HeadGroupSpec.RayStartParams
	if startParams == nil {
		startParams = map[string]string{
			"redis-password": "$REDIS_PASSWORD",
		}
	} else {
		startParams["redis-password"] = "$REDIS_PASSWORD"
	}

	return admission.PatchOp{
		Op:    admission.PatchOpReplace,
		Path:  "/spec/headGroupSpec/rayStartParams",
		Value: startParams,
	}
}

func (m *mutator) Resource() admission.Resource {
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
