package cluster

import (
	"context"
	"fmt"

	ctlv1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/core/v1"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	ctlkuberayv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/ray.io/v1"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
)

// clusterHandler reconcile the user's clusterRole and clusterRoleBinding
type clusterHandler struct {
	rayClusters  ctlkuberayv1.RayClusterClient
	services     ctlv1.ServiceClient
	serviceCache ctlv1.ServiceCache
}

func Register(ctx context.Context, management *config.Management) error {
	clusters := management.KubeRayFactory.Ray().V1().RayCluster()
	services := management.CoreFactory.Core().V1().Service()

	h := &clusterHandler{
		rayClusters:  clusters,
		services:     services,
		serviceCache: services.Cache(),
	}

	clusters.OnChange(ctx, "ob-ray-cluster", h.OnChanged)
	return nil
}

func (h *clusterHandler) OnChanged(_ string, cluster *rayv1.RayCluster) (*rayv1.RayCluster, error) {
	if cluster == nil || cluster.DeletionTimestamp != nil {
		return cluster, nil
	}

	if err := h.ensureService(cluster); err != nil {
		return cluster, err
	}

	return nil, nil
}

// ensureService ensures the service for the ray cluster
func (h *clusterHandler) ensureService(cluster *rayv1.RayCluster) error {
	// check if service already exists by annotation
	if cluster.Annotations == nil {
		return nil
	}
	if _, ok := cluster.Annotations[constant.EnabledExposeSvcAnnotation]; !ok {
		return nil
	}

	svcName := getClusterServiceName(cluster.Name)

	svc, err := h.serviceCache.Get(cluster.Namespace, svcName)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// only create service if not exist
	if svc != nil {
		return nil
	}

	exposeSvc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: cluster.Namespace,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{
				{
					Name:       "client",
					Port:       10001,
					TargetPort: intstr.FromString("client"),
				},
				{
					Name:       "dashboard",
					Port:       8265,
					TargetPort: intstr.FromString("dashboard"),
				},
			},
		},
	}

	exposeSvc.Spec.Selector = map[string]string{
		"ray.io/cluster":    cluster.Name,
		"ray.io/identifier": fmt.Sprintf("%s-head", cluster.Name),
		"ray.io/node-type":  "head",
	}

	if _, err := h.services.Create(exposeSvc); err != nil {
		return err
	}
	return nil
}

func getClusterServiceName(clusterName string) string {
	return fmt.Sprintf("ob-%s-expose", clusterName)
}
