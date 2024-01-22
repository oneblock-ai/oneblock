package cluster

import (
	"context"

	ctlcorev1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/core/v1"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ctlkuberayv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/ray.io/v1"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
)

const (
	kubeRayControllerSyncCluster = "OneblockRayCluster.syncCluster"
	kubeRayControllerCreatePVC   = "OneblockRayCluster.createPVC"
)

// handler reconcile the user's clusterRole and clusterRoleBinding
type handler struct {
	releaseName  string
	rayClusters  ctlkuberayv1.RayClusterClient
	services     ctlcorev1.ServiceClient
	serviceCache ctlcorev1.ServiceCache
	secrets      ctlcorev1.SecretClient
	secretsCache ctlcorev1.SecretCache
	pvcs         ctlcorev1.PersistentVolumeClaimClient
	pvcCache     ctlcorev1.PersistentVolumeClaimCache
}

func Register(ctx context.Context, mgmt *config.Management) error {
	clusters := mgmt.KubeRayFactory.Ray().V1().RayCluster()
	services := mgmt.CoreFactory.Core().V1().Service()
	secrets := mgmt.CoreFactory.Core().V1().Secret()
	pvcs := mgmt.CoreFactory.Core().V1().PersistentVolumeClaim()

	h := &handler{
		releaseName:  mgmt.ReleaseName,
		rayClusters:  clusters,
		services:     services,
		serviceCache: services.Cache(),
		secrets:      secrets,
		secretsCache: secrets.Cache(),
		pvcs:         pvcs,
		pvcCache:     pvcs.Cache(),
	}

	clusters.OnChange(ctx, kubeRayControllerSyncCluster, h.OnChanged)
	clusters.OnChange(ctx, kubeRayControllerCreatePVC, h.createPVCFromAnnotation)
	return nil
}

func (h *handler) OnChanged(_ string, cluster *rayv1.RayCluster) (*rayv1.RayCluster, error) {
	if cluster == nil || cluster.DeletionTimestamp != nil {
		return cluster, nil
	}

	// sync GCS Redis secret to the ray cluster namespace
	h.syncGCSRedisSecretToNamespace(h.releaseName, cluster)

	if err := h.ensureService(cluster); err != nil {
		return cluster, err
	}

	return nil, nil
}

// ensureService create an expose service for the ray cluster if the annotation is specified by the user
func (h *handler) ensureService(cluster *rayv1.RayCluster) error {
	// check if service already exists by annotation
	if cluster.Annotations == nil {
		return nil
	}
	v, ok := cluster.Annotations[constant.AnnotationEnabledExposeSvcKey]
	if !ok {
		return nil
	}

	svcName := getClusterExposeServiceName(cluster.Name)
	svc, err := h.serviceCache.Get(cluster.Namespace, svcName)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if v == "true" && svc == nil {
		newSvc := getExposeSvc(cluster)
		if _, err := h.services.Create(newSvc); err != nil {
			return err
		}
		// only create service if not exist
		if svc != nil {
			return nil
		}

		return err
	} else if v == "false" && svc != nil {
		return h.services.Delete(svc.Namespace, svc.Name, &metav1.DeleteOptions{})
	}

	return nil
}
