package mlservice

import (
	"fmt"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"

	mlv1 "github.com/oneblock-ai/oneblock/pkg/apis/ml.oneblock.ai/v1"
	"github.com/oneblock-ai/oneblock/pkg/controller/kuberay/cluster"
	"github.com/oneblock-ai/oneblock/pkg/utils"
)

const (
	defaultRayLLMVersion       = "0.5.0"
	defaultRayLLMImage         = "anyscale/ray-ml"
	huggingFaceHubTokenEnvName = "HUGGING_FACE_HUB_TOKEN" // #nosec G101
)

func GetRayClusterSpecConfig(mlSvc *mlv1.MLService, modelTmpVersion *mlv1.ModelTemplateVersion, releaseName string) (*rayv1.RayClusterSpec, error) {
	mlClusterRef := mlSvc.Spec.MLClusterRef
	clusterImg, version := getRayClusterImage(mlClusterRef.RayClusterSpec.Image, mlClusterRef.RayClusterSpec.Version)

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

	rayClusterSpec := &rayv1.RayClusterSpec{
		RayVersion:       version,
		HeadGroupSpec:    headGroupSpec,
		WorkerGroupSpecs: workerGroupSpecs,
	}

	// autoscaling is enabled by default
	if mlClusterRef.RayClusterSpec.EnableAutoScaling {
		rayClusterSpec.AutoscalerOptions, err = getDefaultAutoScalingOptions()
		if err != nil {
			return nil, err
		}
	}

	return rayClusterSpec, nil
}

func getDefaultAutoScalingOptions() (*rayv1.AutoscalerOptions, error) {
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

	return &rayv1.AutoscalerOptions{
		UpscalingMode:      &upScalingMode,
		IdleTimeoutSeconds: pointer.Int32(60),
		ImagePullPolicy:    &pullPolicy,
		Resources: &corev1.ResourceRequirements{
			Requests: scaleRequests,
			Limits:   scaleLimits,
		},
	}, nil
}

// GetHeadGroupSpecConfig returns the head group spec of the rayCluster
// 1. GCS and persistent log is enabled by default for the head group
// 2. add model config mount point
func GetHeadGroupSpecConfig(mlService *mlv1.MLService, modelTmpVersion *mlv1.ModelTemplateVersion, releaseName, image string) (rayv1.HeadGroupSpec, error) {
	headGroupSpec := rayv1.HeadGroupSpec{}

	defaultStartParams := map[string]string{
		"num-cpus":       "0", // Setting "num-cpus: 0" to avoid any Ray actors or tasks being scheduled on the Ray head Pod.
		"redis-password": "$REDIS_PASSWORD",
		"dashboard-host": "0.0.0.0",
	}
	rayStartParams := mlService.Spec.MLClusterRef.RayClusterSpec.HeadGroupSpec.RayStartParams
	if rayStartParams != nil {
		for k, v := range rayStartParams {
			defaultStartParams[k] = v
		}
	}
	headGroupSpec.ServiceType = mlService.Spec.MLClusterRef.RayClusterSpec.HeadGroupSpec.ServiceType
	headGroupSpec.RayStartParams = defaultStartParams

	var requests, limits corev1.ResourceList
	var err error

	mlHeadGroupSpec := mlService.Spec.MLClusterRef.RayClusterSpec.HeadGroupSpec
	if mlHeadGroupSpec.Resources == nil {
		requests, err = getResourceList("500m", "1Gi")
		if err != nil {
			return headGroupSpec, err
		}

		limits, err = getResourceList("1", "2Gi")
		if err != nil {
			return headGroupSpec, err
		}
	} else {
		requests = mlHeadGroupSpec.Resources.Requests
		limits = mlHeadGroupSpec.Resources.Limits
	}

	// config pod template for head node
	podTemplate := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "ray-head",
					Image: image,
					Ports: getDefaultClusterPorts(),
					Env:   cluster.GetHeadNodeRedisEnvConfig(releaseName, mlService.Namespace),
					Resources: corev1.ResourceRequirements{
						Requests: requests,
						Limits:   limits,
					},
					Lifecycle: getClusterDefaultLifecycle(),
				},
			},
		},
	}

	if mlHeadGroupSpec.Volume != nil {
		podTemplate.Spec.Volumes = []corev1.Volume{
			{
				Name: "ray-logs",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: mlHeadGroupSpec.Volume.Name,
					},
				},
			},
		}
		podTemplate.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "ray-logs",
				MountPath: "/tmp/ray",
			},
		}
	}

	// add model config
	if mlService.Spec.ModelTemplateVersionRef != nil {
		modelVol := GetModelVolume(modelTmpVersion)
		podTemplate.Spec.Volumes = append(podTemplate.Spec.Volumes, modelVol)
		podTemplate.Spec.Containers[0].VolumeMounts = append(podTemplate.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      "model",
			MountPath: "/home/ray/models",
		})
	}

	headGroupSpec.Template = podTemplate
	return headGroupSpec, nil
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

	podTemplate := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:      "ray-worker",
					Image:     image,
					Lifecycle: getClusterDefaultLifecycle(),
					Env:       workNodeEnv,
				},
			},
		},
	}

	if wgCfg.Resources != nil {
		podTemplate.Spec.Containers[0].Resources = *wgCfg.Resources
	}

	if wgCfg.RuntimeClassName != nil {
		podTemplate.Spec.RuntimeClassName = wgCfg.RuntimeClassName
	}

	if wgCfg.NodeSelector != nil && len(wgCfg.NodeSelector) > 0 {
		podTemplate.Spec.NodeSelector = wgCfg.NodeSelector
	}

	if wgCfg.Tolerations != nil && len(wgCfg.Tolerations) > 0 {
		podTemplate.Spec.Tolerations = wgCfg.Tolerations
	}

	if wgCfg.Volume != nil {
		podTemplate.Spec.Volumes = []corev1.Volume{
			{
				Name: wgCfg.Volume.Name,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: wgCfg.Volume.Name,
					},
				},
			},
		}
		podTemplate.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      wgCfg.Volume.Name,
				MountPath: "/home/ray/.cache",
			},
		}
	}

	workerGroupSpec.Template = podTemplate

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

func getRayClusterImage(image, version string) (string, string) {
	if image == "" {
		image = defaultRayLLMImage
	}

	if version == "" {
		version = defaultRayLLMVersion
	}

	return fmt.Sprintf("%s:%s", image, version), version
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
			Name:          "mlserve",
			ContainerPort: 8000,
		},
	}
}

func SetRayClusterImage(mlSvc *mlv1.MLService, service *rayv1.RayService) {
	img, version := getRayClusterImage(mlSvc.Spec.MLClusterRef.RayClusterSpec.Image, mlSvc.Spec.MLClusterRef.RayClusterSpec.Version)
	service.Spec.RayClusterSpec.RayVersion = version
	service.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.Containers[0].Image = img
	service.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.Spec.Containers[0].Image = img
}

func SetRayClusterHeadConfig(mlSvc *mlv1.MLService, service *rayv1.RayService) {
	headGroupConfig := mlSvc.Spec.MLClusterRef.RayClusterSpec.HeadGroupSpec
	service.Spec.RayClusterSpec.HeadGroupSpec.ServiceType = headGroupConfig.ServiceType
	for k, v := range headGroupConfig.RayStartParams {
		service.Spec.RayClusterSpec.HeadGroupSpec.RayStartParams[k] = v
	}
	if headGroupConfig.Resources != nil {
		service.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.Containers[0].Resources = *headGroupConfig.Resources
	}
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
