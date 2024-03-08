package mlservice

import (
	"fmt"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"

	mlv1 "github.com/oneblock-ai/oneblock/pkg/apis/ml.oneblock.ai/v1"
	"github.com/oneblock-ai/oneblock/pkg/controller/raycluster"
	"github.com/oneblock-ai/oneblock/pkg/settings"
	"github.com/oneblock-ai/oneblock/pkg/utils"
)

const (
	huggingFaceHubTokenEnvName = "HUGGING_FACE_HUB_TOKEN" // #nosec G101
)

func GetRayClusterSpecConfig(mlSvc *mlv1.MLService, modelTmpVersion *mlv1.ModelTemplateVersion, releaseName string) (*rayv1.RayClusterSpec, error) {
	mlClusterRef := mlSvc.Spec.MLClusterRef
	clusterImg := getRayClusterImage(mlClusterRef.RayClusterSpec.Image)

	headGroupSpec, err := GetHeadGroupSpecConfig(mlSvc, modelTmpVersion, releaseName, clusterImg)
	if err != nil {
		return nil, err
	}

	workerGroupSpecs := make([]rayv1.WorkerGroupSpec, len(mlSvc.Spec.MLClusterRef.RayClusterSpec.WorkerGroupSpec))
	for i, wgCfg := range mlSvc.Spec.MLClusterRef.RayClusterSpec.WorkerGroupSpec {
		workerGroupSpecs[i], err = GetDefaultWorkerGroupSpecConfig(wgCfg, clusterImg, mlSvc.Spec.HFSecretRef)
		if err != nil {
			return nil, err
		}
	}

	return &rayv1.RayClusterSpec{
		HeadGroupSpec:     *headGroupSpec,
		WorkerGroupSpecs:  workerGroupSpecs,
		AutoscalerOptions: getDefaultAutoScalingOptions(),
	}, nil
}

func getDefaultAutoScalingOptions() *rayv1.AutoscalerOptions {
	upScalingMode := rayv1.UpscalingMode("Default")
	pullPolicy := corev1.PullIfNotPresent

	return &rayv1.AutoscalerOptions{
		UpscalingMode:      &upScalingMode,
		IdleTimeoutSeconds: pointer.Int32(60),
		ImagePullPolicy:    &pullPolicy,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("200m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
		},
	}
}

// GetHeadGroupSpecConfig returns the head group spec of the rayCluster
// 1. GCS and persistent log is enabled by default for the head group
// 2. add model config mount point
func GetHeadGroupSpecConfig(mlService *mlv1.MLService, modelTmpVersion *mlv1.ModelTemplateVersion, releaseName, image string) (*rayv1.HeadGroupSpec, error) {
	rayStartParams := map[string]string{
		"num-cpus":       "0", // Setting "num-cpus: 0" to avoid any Ray actors or tasks being scheduled on the Ray head Pod.
		"redis-password": "$REDIS_PASSWORD",
		"dashboard-host": "0.0.0.0",
		"block":          "true",
	}

	headVolumes := []corev1.Volume{
		{
			Name: "ray-logs",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: getHeadGroupVolName(mlService.Name),
				},
			},
		},
	}

	// config pod template for head node
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "ray-head",
				Image: image,
				Ports: getDefaultClusterPorts(),
				Env:   raycluster.GetHeadNodeRedisEnvConfig(releaseName, mlService.Namespace),
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
				Lifecycle: getClusterDefaultLifecycle(),
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "ray-logs",
						MountPath: "/tmp/ray",
					},
				},
			},
		},
		Volumes: headVolumes,
	}

	// add model config
	if mlService.Spec.ModelTemplateVersionRef != nil {
		modelVol := GetModelVolume(modelTmpVersion)
		podSpec.Volumes = append(podSpec.Volumes, modelVol)
		podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      "model",
			MountPath: "/home/ray/models",
		})
	}

	return &rayv1.HeadGroupSpec{
		RayStartParams: rayStartParams,
		Template: corev1.PodTemplateSpec{
			Spec: podSpec,
		},
	}, nil
}

func GetModelVolume(modelTmpVersion *mlv1.ModelTemplateVersion) corev1.Volume {
	return corev1.Volume{
		Name: "model",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: modelTmpVersion.Name,
				},
				Items: []corev1.KeyToPath{
					{
						Key:  GetModelConfigMapKey(modelTmpVersion.Name),
						Path: GetModelConfigMapKey(modelTmpVersion.Name),
					},
				},
			},
		},
	}
}

func GetDefaultWorkerGroupSpecConfig(wgCfg mlv1.WorkerGroupSpec, image string, hfRef *mlv1.HFSecretRef) (rayv1.WorkerGroupSpec, error) {
	workerGroupSpec := rayv1.WorkerGroupSpec{}
	workerGroupSpec.GroupName = wgCfg.Name
	workerGroupSpec.Replicas = wgCfg.Replicas
	workerGroupSpec.MinReplicas = wgCfg.MinReplicas
	workerGroupSpec.MaxReplicas = wgCfg.MaxReplicas

	if wgCfg.RayStartParams != nil {
		workerGroupSpec.RayStartParams = wgCfg.RayStartParams
	} else {
		workerGroupSpec.RayStartParams = map[string]string{
			"block": "true",
		}
	}

	if wgCfg.AcceleratorTypes != nil && len(wgCfg.AcceleratorTypes) > 0 {
		workerGroupSpec.RayStartParams["resources"] = configResourceAccelerators(wgCfg)
	}

	workNodeEnv := []corev1.EnvVar{
		{
			Name:  "RAY_gcs_rpc_server_reconnect_timeout_s",
			Value: "300",
		},
	}

	// the SyncClusterSecretsToLocalNS method will helps to sync the HF secret to the local namespace
	if hfRef != nil {
		workNodeEnv = append(workNodeEnv, corev1.EnvVar{
			Name: huggingFaceHubTokenEnvName,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: hfRef.Name,
					},
					Key: hfRef.SecretKey,
				},
			},
		})
	}

	podTemplateSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:      "ray-worker",
				Image:     image,
				Lifecycle: getClusterDefaultLifecycle(),
				Env:       workNodeEnv,
			},
		},
	}

	if wgCfg.Resources != nil {
		podTemplateSpec.Containers[0].Resources = *wgCfg.Resources
	}

	if wgCfg.RuntimeClassName != nil {
		podTemplateSpec.RuntimeClassName = wgCfg.RuntimeClassName
	}

	if wgCfg.NodeSelector != nil && len(wgCfg.NodeSelector) > 0 {
		podTemplateSpec.NodeSelector = wgCfg.NodeSelector
	}

	if wgCfg.Tolerations != nil && len(wgCfg.Tolerations) > 0 {
		podTemplateSpec.Tolerations = wgCfg.Tolerations
	}

	if wgCfg.Volume != nil {
		podTemplateSpec.Volumes = []corev1.Volume{
			{
				Name: wgCfg.Volume.Name,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: wgCfg.Volume.Name,
					},
				},
			},
		}
		podTemplateSpec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      wgCfg.Volume.Name,
				MountPath: "/home/ray/.cache",
			},
		}
	}

	workerGroupSpec.Template.Spec = podTemplateSpec

	return workerGroupSpec, nil
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
		corev1.ResourceCPU:    cpuRes,
		corev1.ResourceMemory: memRes,
	}, nil
}

func getRayClusterImage(image string) string {
	if image != "" {
		return image
	}
	return settings.RayLLMImage.Get()
}

func getClusterDefaultLifecycle() *corev1.Lifecycle {
	return &corev1.Lifecycle{
		PreStop: &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"/bin/sh", "-c", "ray stop"},
			},
		},
	}

}

func getDefaultClusterPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "gcs-server",
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
	}
}

func SetRayClusterImage(mlSvc *mlv1.MLService, service *rayv1.RayService) {
	img := getRayClusterImage(mlSvc.Spec.MLClusterRef.RayClusterSpec.Image)
	service.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.Containers[0].Image = img
	service.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec.Containers[0].Image = img
}

func SetRayClusterWorkerGroupConfig(mlSvc *mlv1.MLService, service *rayv1.RayService) {
	for _, workerGroup := range mlSvc.Spec.MLClusterRef.RayClusterSpec.WorkerGroupSpec {
		for i, svcWorkerGroup := range service.Spec.RayClusterSpec.WorkerGroupSpecs {
			if svcWorkerGroup.GroupName == workerGroup.Name {
				svcWorkerGroup.Replicas = workerGroup.Replicas
				svcWorkerGroup.MinReplicas = workerGroup.MinReplicas
				svcWorkerGroup.MaxReplicas = workerGroup.MaxReplicas
				for k, v := range workerGroup.RayStartParams {
					svcWorkerGroup.RayStartParams[k] = v
				}

				if workerGroup.AcceleratorTypes != nil && len(workerGroup.AcceleratorTypes) > 0 {
					svcWorkerGroup.RayStartParams["resources"] = configResourceAccelerators(workerGroup)
				}

				svcWorkerGroup.Template.Spec.RuntimeClassName = workerGroup.RuntimeClassName
				svcWorkerGroup.Template.Spec.NodeSelector = workerGroup.NodeSelector
				svcWorkerGroup.Template.Spec.Tolerations = workerGroup.Tolerations
				if workerGroup.Resources != nil {
					svcWorkerGroup.Template.Spec.Containers[0].Resources = *workerGroup.Resources
				}

				hfRef := mlSvc.Spec.HFSecretRef
				workerGroupEnv := setWorkerGroupHFEnv(hfRef, svcWorkerGroup)
				if workerGroupEnv != nil {
					svcWorkerGroup.Template.Spec.Containers[0].Env = workerGroupEnv
				}
				service.Spec.RayClusterSpec.WorkerGroupSpecs[i] = svcWorkerGroup
				break
			}
		}
	}
}

// setWorkerGroupHFEnv sets the Hugging Face Hub token env for the worker group
func setWorkerGroupHFEnv(hfRef *mlv1.HFSecretRef, workerGroup rayv1.WorkerGroupSpec) []corev1.EnvVar {
	if hfRef == nil {
		return nil
	}
	workerGroupEnv := workerGroup.Template.Spec.Containers[0].Env
	for _, env := range workerGroupEnv {
		if env.Name == huggingFaceHubTokenEnvName && env.ValueFrom.SecretKeyRef.Name == hfRef.Name {
			return nil
		}
	}
	workerGroupEnv = append(workerGroupEnv, corev1.EnvVar{
		Name: huggingFaceHubTokenEnvName,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: hfRef.Name,
				},
				Key: hfRef.SecretKey,
			},
		},
	})
	return workerGroupEnv
}

func configResourceAccelerators(spec mlv1.WorkerGroupSpec) string {
	var accelerators string
	if spec.AcceleratorTypes != nil {
		for aType, v := range spec.AcceleratorTypes {
			acceleratorType := utils.GetAcceleratorTypeByProductName(aType)
			accelerators += fmt.Sprintf("\"accelerator_type:%s\": %d,", acceleratorType, v)
		}
		return fmt.Sprintf("\"{%s}\"", accelerators[:len(accelerators)-1])
	}
	return ""
}

func getHeadGroupVolName(mlSvcName string) string {
	return fmt.Sprintf(mlSvcName + "-head-log")
}

func getHeadGroupVolume(mlSvc *mlv1.MLService) mlv1.Volume {
	return mlv1.Volume{
		Name: getHeadGroupVolName(mlSvc.Name),
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("5Gi"),
				},
			},
		},
	}
}
