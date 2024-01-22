package cluster

import (
	"fmt"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getExposeSvc(cluster *rayv1.RayCluster) *v1.Service {
	selector := map[string]string{
		"ray.io/cluster":    cluster.Name,
		"ray.io/identifier": fmt.Sprintf("%s-head", cluster.Name),
		"ray.io/node-type":  "head",
	}
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getClusterExposeServiceName(cluster.Name),
			Namespace: cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: cluster.APIVersion,
					Kind:       cluster.Kind,
					Name:       cluster.Name,
					UID:        cluster.UID,
				},
			},
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
			Selector: selector,
		},
	}
}

func getClusterExposeServiceName(name string) string {
	return fmt.Sprintf("%s-expose", name)
}
