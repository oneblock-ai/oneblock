package mlservice

import (
	"fmt"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mlv1 "github.com/oneblock-ai/oneblock/pkg/apis/ml.oneblock.ai/v1"
	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
)

type ServeConfig struct {
	Applications []ServeApplication `yaml:"applications,omitempty"`
}

type ServeApplication struct {
	Name        string    `yaml:"name,omitempty"`
	RoutePrefix string    `yaml:"route_prefix,omitempty"`
	ImportPath  string    `yaml:"import_path,omitempty"`
	Args        ServeArgs `yaml:"args,omitempty"`
}

type ServeArgs struct {
	Models []string `json:"models,omitempty"`
}

func getRayServiceConfig(mlService *mlv1.MLService, modelTmpVersion *mlv1.ModelTemplateVersion,
	owners []metav1.OwnerReference, releaseName string) (*rayv1.RayService, error) {

	serveConfig, err := getServeConfigV2(mlService.Name, getModelConfigPath(modelTmpVersion.Name))
	if err != nil {
		return nil, err
	}

	rayClusterSpec, err := GetRayClusterSpecConfig(mlService, modelTmpVersion, releaseName)
	if err != nil {
		return nil, err
	}

	raySvc := &rayv1.RayService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mlService.Name,
			Namespace: mlService.Namespace,
			Annotations: map[string]string{
				constant.AnnotationRayFTEnabledKey:    "true",
				constant.AnnoModelTemplateVersionName: modelTmpVersion.Name,
			},
			OwnerReferences: owners,
		},
		Spec: rayv1.RayServiceSpec{
			ServeConfigV2:  serveConfig,
			RayClusterSpec: *rayClusterSpec,
		},
	}

	return raySvc, nil
}

func getModelConfigPath(modelName string) string {
	return fmt.Sprintf("./models/%s.yaml", modelName)
}

func getServeConfigV2(name, modelPath string) (string, error) {
	serveCfg := &ServeConfig{
		Applications: []ServeApplication{
			{
				Name:        name,
				RoutePrefix: "/",
				ImportPath:  "rayllm.backend:router_application",
				Args: ServeArgs{
					Models: []string{
						modelPath,
					},
				},
			},
		},
	}
	serveCfgStr, err := yaml.Marshal(serveCfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal mlserve config: %v", err)
	}
	return string(serveCfgStr), nil
}

func GetModelConfigMapKey(modelTpmVersionName string) string {
	return fmt.Sprintf("%s.yaml", modelTpmVersionName)
}
