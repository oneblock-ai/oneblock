package data

import (
	"context"
	"encoding/json"

	"github.com/rancher/wrangler/v2/pkg/apply"
	ctlcorev1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/core/v1"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/oneblock-ai/oneblock/pkg/controller/raycluster"
	ctlrayv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/ray.io/v1"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
	"github.com/oneblock-ai/oneblock/pkg/settings"
	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
)

const (
	defaultRayClusterName = "default-cluster"
	defaultRayPVCName     = defaultRayClusterName + "-log"
	defaultQueue          = "default"
)

type handler struct {
	ctx         context.Context
	apply       apply.Apply
	namespaces  ctlcorev1.NamespaceClient
	rayClusters ctlrayv1.RayClusterClient
}

func addDefaultPublicRayCluster(ctx context.Context, mgmt *config.Management, name string) error {
	// add default public ray raycluster for all authenticated users
	nss := mgmt.CoreFactory.Core().V1().Namespace()
	rayClusters := mgmt.KubeRayFactory.Ray().V1().RayCluster()
	h := &handler{
		ctx:         ctx,
		apply:       mgmt.Apply,
		namespaces:  nss,
		rayClusters: rayClusters,
	}

	// check if the default ray raycluster has been initialized by ns annotation
	ns, err := h.namespaces.Get(constant.PublicNamespaceName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if ns.Annotations != nil && ns.Annotations[constant.AnnotationRayClusterInitialized] == "true" {
		logrus.Infof("Skipping creating default ray raycluster %s, it has already been initialized", defaultRayClusterName)
		return nil
	}

	rayCluster, err := getDefaultRayCluster(name)
	if err != nil {
		return err
	}

	if _, err := h.rayClusters.Create(rayCluster); err != nil {
		return err
	}

	nsCpy := ns.DeepCopy()
	if nsCpy.Annotations == nil {
		nsCpy.Annotations = map[string]string{}
	}
	nsCpy.Annotations[constant.AnnotationRayClusterInitialized] = "true"

	if _, err = h.namespaces.Update(nsCpy); ns != nil {
		return err
	}

	return nil
}

func getDefaultRayCluster(releaseName string) (*rayv1.RayCluster, error) {
	pvcAnno, err := getDefaultLogPVC()
	if err != nil {
		return nil, err
	}

	annotations := map[string]string{
		constant.AnnotationRayFTEnabledKey:      "true", // enable GCS fault tolerance
		constant.AnnotationVolumeClaimTemplates: string(pvcAnno),
	}

	labels := map[string]string{
		constant.LabelRaySchedulerName: constant.VolcanoSchedulerName,
		constant.LabelVolcanoQueueName: defaultQueue,
	}

	rayStartParams := map[string]string{
		"num-cpus":       "0", // Setting "num-cpus: 0" to avoid any Ray actors or tasks being scheduled on the Ray head Pod.
		"redis-password": "$REDIS_PASSWORD",
		"dashboard-host": "0.0.0.0",
		"block":          "true",
	}

	return &rayv1.RayCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:        defaultRayClusterName,
			Namespace:   constant.PublicNamespaceName,
			Annotations: annotations,
			Labels:      labels,
		},
		Spec: rayv1.RayClusterSpec{
			RayVersion:              settings.RayVersion.Get(),
			EnableInTreeAutoscaling: pointer.Bool(true),
			HeadGroupSpec: rayv1.HeadGroupSpec{
				RayStartParams: rayStartParams,
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "ray-head",
								Image: settings.RayClusterImage.Get(),
								Ports: []corev1.ContainerPort{
									{
										Name:          "gcs",
										ContainerPort: 6379,
									},
									{
										Name:          "client",
										ContainerPort: 10001,
									},
									{
										Name:          "dashboard",
										ContainerPort: 8265,
									},
									{
										Name:          "serve",
										ContainerPort: 8000,
									},
								},
								Env: raycluster.GetHeadNodeRedisEnvConfig(releaseName, constant.PublicNamespaceName),
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("500m"),
										corev1.ResourceMemory: resource.MustParse("1Gi"),
									},
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("1"),
										corev1.ResourceMemory: resource.MustParse("2Gi"),
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "ray-logs",
										MountPath: "/tmp/ray",
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "ray-logs",
								VolumeSource: corev1.VolumeSource{
									PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
										ClaimName: defaultRayPVCName,
									},
								},
							},
						},
					},
				},
			},
			WorkerGroupSpecs: []rayv1.WorkerGroupSpec{
				{
					Replicas:       pointer.Int32(0), // set default to 0 to save resources since it will be auto-scaled
					MinReplicas:    pointer.Int32(0),
					MaxReplicas:    pointer.Int32(10),
					GroupName:      "default-worker",
					RayStartParams: map[string]string{},
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "ray-worker",
									Image: settings.RayClusterImage.Get(),
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("2"),
											corev1.ResourceMemory: resource.MustParse("4Gi"),
										},
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("2"),
											corev1.ResourceMemory: resource.MustParse("4Gi"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func getDefaultLogPVC() ([]byte, error) {
	pvcs := []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: defaultRayPVCName,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					"ReadWriteOnce",
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"storage": resource.MustParse("5Gi"),
					},
				},
			},
		},
	}

	return json.Marshal(pvcs)
}
