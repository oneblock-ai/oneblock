package data

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rancher/wrangler/v2/pkg/apply"
	ctlcorev1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/core/v1"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/oneblock-ai/oneblock/pkg/controller/kuberay/cluster"
	ctlrayv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/ray.io/v1"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
	"github.com/oneblock-ai/oneblock/pkg/settings"
	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
)

const (
	defaultRayClusterName = "public-ray-cluster"
	defaultRayPVCName     = defaultRayClusterName + "-log"
)

type handler struct {
	ctx         context.Context
	apply       apply.Apply
	namespaces  ctlcorev1.NamespaceClient
	rayClusters ctlrayv1.RayClusterClient
}

func addDefaultPublicRayCluster(ctx context.Context, mgmt *config.Management, name string) error {
	// add default public ray cluster for all authenticated users
	nss := mgmt.CoreFactory.Core().V1().Namespace()
	rayClusters := mgmt.KubeRayFactory.Ray().V1().RayCluster()
	h := &handler{
		ctx:         ctx,
		apply:       mgmt.Apply,
		namespaces:  nss,
		rayClusters: rayClusters,
	}

	// check if the default ray cluster has been initialized by ns annotation
	ns, err := h.namespaces.Get(constant.PublicNamespaceName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if ns.Annotations != nil && ns.Annotations[constant.AnnotationRayClusterInitialized] == "true" {
		logrus.Infof("Skipping creating default ray cluster %s, it has already been initialized", defaultRayClusterName)
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

func getDefaultLogPVC() (string, error) {
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
						"storage": resource.MustParse("1Gi"),
					},
				},
			},
		},
	}

	pvcByte, err := json.Marshal(pvcs)
	if err != nil {
		return "", err
	}
	return string(pvcByte), nil
}

func getResourceList(cpu, memory string) (corev1.ResourceList, error) {
	cpuRes, err := resource.ParseQuantity(cpu)
	if err != nil {
		return nil, err

	}
	memRes, err := resource.ParseQuantity(memory)
	if err != nil {
		return nil, err
	}

	return corev1.ResourceList{
		"cpu":    cpuRes,
		"memory": memRes,
	}, nil
}

func getDefaultRayCluster(releaseName string) (*rayv1.RayCluster, error) {
	upScalingMode := rayv1.UpscalingMode("Default")
	pullPolicy := corev1.PullIfNotPresent
	scaleRequests, err := getResourceList("200m", "256Mi")
	if err != nil {
		return nil, err
	}

	scaleLimits, err := getResourceList("500m", "512Mi")
	if err != nil {
		return nil, err
	}

	headResourceReq, err := getResourceList("500m", "1Gi")
	if err != nil {
		return nil, err
	}

	headResourceLim, err := getResourceList("1", "2Gi")
	if err != nil {
		return nil, err
	}

	workerResourceReq, err := getResourceList("1", "2Gi")
	if err != nil {
		return nil, err
	}

	workerResourceLim, err := getResourceList("2", "4Gi")
	if err != nil {
		return nil, err
	}

	pvcAnno, err := getDefaultLogPVC()
	if err != nil {
		return nil, err
	}

	annotations := map[string]string{
		constant.AnnotationRayFTEnabledKey:      "true", // enable Ray GCS FT
		constant.AnnotationEnabledExposeSvcKey:  "true", // auto generate exposed svc
		constant.AnnotationVolumeClaimTemplates: pvcAnno,
	}

	rayStartParams := map[string]string{
		"num-cpus":       "0", // Setting "num-cpus: 0" to avoid any Ray actors or tasks being scheduled on the Ray head Pod.
		"redis-password": "$REDIS_PASSWORD",
	}

	workNodeEnv := []corev1.EnvVar{
		{
			Name:  "RAY_gcs_rpc_server_reconnect_timeout_s",
			Value: "300",
		},
	}

	lifecycle := &corev1.Lifecycle{
		PreStop: &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"/bin/sh", "-c", "ray stop"},
			},
		},
	}

	return &rayv1.RayCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:        defaultRayClusterName,
			Namespace:   constant.PublicNamespaceName,
			Annotations: annotations,
			Labels: map[string]string{
				constant.LabelRaySchedulerName: constant.VolcanoSchedulerName,
				constant.LabelVolcanoQueueName: defaultQueueName,
			},
		},
		Spec: rayv1.RayClusterSpec{
			RayVersion:              settings.RayVersion.Get(),
			EnableInTreeAutoscaling: pointer.Bool(true),
			AutoscalerOptions: &rayv1.AutoscalerOptions{
				UpscalingMode:      &upScalingMode,
				IdleTimeoutSeconds: pointer.Int32(60),
				ImagePullPolicy:    &pullPolicy,
				Resources: &corev1.ResourceRequirements{
					Requests: scaleRequests,
					Limits:   scaleLimits,
				},
			},
			HeadGroupSpec: rayv1.HeadGroupSpec{
				RayStartParams: rayStartParams,
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "ray-head",
								Image: fmt.Sprintf("rayproject/ray:%s", settings.RayVersion.Get()),
								Ports: []corev1.ContainerPort{
									{
										Name:          "redis",
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
								},
								Env: cluster.GetHeadNodeRedisEnvConfig(releaseName, constant.PublicNamespaceName),
								Resources: corev1.ResourceRequirements{
									Requests: headResourceReq,
									Limits:   headResourceLim,
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "ray-logs",
										MountPath: "/tmp/ray",
									},
								},
								Lifecycle: lifecycle,
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
					Replicas:       pointer.Int32(1),
					MinReplicas:    pointer.Int32(1),
					MaxReplicas:    pointer.Int32(10),
					GroupName:      "default-worker",
					RayStartParams: map[string]string{},
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "ray-worker",
									Image: fmt.Sprintf("rayproject/ray:%s", settings.RayVersion.Get()),
									Resources: corev1.ResourceRequirements{
										Requests: workerResourceReq,
										Limits:   workerResourceLim,
									},
									Lifecycle: lifecycle,
									Env:       workNodeEnv,
								},
							},
						},
					},
				},
			},
		},
	}, nil
}
