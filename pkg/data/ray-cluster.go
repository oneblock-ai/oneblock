package data

import (
	"context"
	"fmt"
	"time"

	"github.com/rancher/wrangler/v2/pkg/apply"
	ctlcorev1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/core/v1"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	ctlrayv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/ray.io/v1"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
	"github.com/oneblock-ai/oneblock/pkg/settings"
	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
)

const (
	defaultRayPVCName     = "ray-tmp-logs"
	defaultRayClusterName = "public-ray-cluster"
)

type hander struct {
	ctx          context.Context
	apply        apply.Apply
	secretsCache ctlcorev1.SecretCache
	rayCluster   ctlrayv1.RayClusterClient
}

func addDefaultPublicRayCluster(ctx context.Context, mgmt *config.Management, name string) error {
	// add default public ray cluster for all authenticated users
	handler := &hander{
		ctx:          ctx,
		apply:        mgmt.Apply,
		secretsCache: mgmt.CoreFactory.Core().V1().Secret().Cache(),
		rayCluster:   mgmt.KubeRayFactory.Ray().V1().RayCluster(),
	}

	// skip if default ray cluster is already created
	ray, err := handler.rayCluster.Get(defaultPublicNamespace, defaultRayClusterName, metav1.GetOptions{})
	if ray != nil && err == nil {
		logrus.Infof("skip creating default ray cluster %s:%s, already exist", ray.Namespace, ray.Name)
		return nil
	}

	// create default ray log pvc
	if err := handler.createRayLogPVC(); err != nil {
		logrus.Errorf("failed to create ray default pvc: %v", err)
		return err
	}

	handler.waitForRedisSecret(name)

	upScalingMode := rayv1.UpscalingMode("Default")
	pullPolicy := corev1.PullIfNotPresent
	scalerRequests, err := getResourceList("200m", "128Mi")
	if err != nil {
		return err
	}

	scalerLimits, err := getResourceList("500m", "512Mi")
	if err != nil {
		return err
	}

	headRequests, err := getResourceList("250m", "512Mi")
	if err != nil {
		return err
	}

	headLimits, err := getResourceList("1", "2G")
	if err != nil {
		return err
	}

	workerRequests, err := getResourceList("1", "1G")
	if err != nil {
		return err
	}

	workerLimits, err := getResourceList("4", "10G")
	if err != nil {
		return err
	}

	annotations := map[string]string{
		"ray.io/ft-enabled":                 "true", // enable Ray GCS FT
		constant.EnabledExposeSvcAnnotation: "true", // auto generate exposed
	}

	rayStartParams := map[string]string{
		"num-cpus":       "0", // Setting "num-cpus: 0" to avoid any Ray actors or tasks being scheduled on the Ray head Pod.
		"redis-password": "$REDIS_PASSWORD",
	}

	headNodeEnv := []corev1.EnvVar{
		{
			Name:  "RAY_REDIS_ADDRESS",
			Value: getRedisSVCName(name),
		},
		{
			Name: "REDIS_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: constant.RedisSecretName,
					},
					Key: constant.RedisSecretKeyName,
				},
			},
		},
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

	rayObj := &rayv1.RayCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "public-ray-cluster",
			Namespace:   defaultPublicNamespace,
			Annotations: annotations,
		},
		Spec: rayv1.RayClusterSpec{
			RayVersion:              settings.RayVersion.Get(),
			EnableInTreeAutoscaling: pointer.Bool(true),
			AutoscalerOptions: &rayv1.AutoscalerOptions{
				UpscalingMode:      &upScalingMode,
				IdleTimeoutSeconds: pointer.Int32(60),
				ImagePullPolicy:    &pullPolicy,
				Resources: &corev1.ResourceRequirements{
					Requests: scalerRequests,
					Limits:   scalerLimits,
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
								Env: headNodeEnv,
								Resources: corev1.ResourceRequirements{
									Requests: headRequests,
									Limits:   headLimits,
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
										Requests: workerRequests,
										Limits:   workerLimits,
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
	}

	return handler.apply.
		WithDynamicLookup().
		WithSetID("public-ray-cluster").
		ApplyObjects(rayObj)
}

func (h *hander) createRayLogPVC() error {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultRayPVCName,
			Namespace: defaultPublicNamespace,
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
	}

	return h.apply.WithDynamicLookup().WithSetID("public-ray-pvc").ApplyObjects(pvc)
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

func (h *hander) waitForRedisSecret(releaseName string) {
	logrus.Debugf("wait for redis secret")
	for {
		// get redis secret
		// if secret is not ready, sleep 2 second
		// if secret is ready, return
		select {
		case <-time.After(2 * time.Second):
			secret, err := h.secretsCache.Get(constant.DefaultSystemNamespace, fmt.Sprintf("%s-redis", releaseName))
			if err != nil {
				logrus.Warnf("Retrying to get redis secret: %v", err)
				continue
			}

			// sync secret to the public namespace
			syncSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constant.RedisSecretName,
					Namespace: defaultPublicNamespace,
				},
				Data: map[string][]byte{
					constant.RedisSecretKeyName: secret.Data[constant.RedisSecretKeyName],
				},
			}

			err = h.apply.WithDynamicLookup().WithSetID("sync-redis-secret-to-public").ApplyObjects(syncSecret)
			if err != nil {
				logrus.Errorf("Retrying to apply redis secret to ns %s: %v", defaultPublicNamespace, err)
				continue
			}
			return
		case <-h.ctx.Done():
			return
		}
	}
}

func getRedisSVCName(releaseName string) string {
	return fmt.Sprintf("redis://%s-redis-master.%s.svc.cluster.local:6379", releaseName, constant.DefaultSystemNamespace)
}
