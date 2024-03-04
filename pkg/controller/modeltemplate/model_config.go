package modeltemplate

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v3"

	mlv1 "github.com/oneblock-ai/oneblock/pkg/apis/ml.oneblock.ai/v1"
)

type RayLLMModelConfig struct {
	DeploymentConfig DeploymentConfig `yaml:"deployment_config"`
	EngineConfig     EngineConfig     `yaml:"engine_config"`
	ScalingConfig    ScalingConfig    `yaml:"scaling_config"`
}

type ScalingConfig struct {
	NumWorkers         int32             `yaml:"num_workers"`
	NumGPUsPerWorker   int32             `yaml:"num_gpus_per_worker"`
	NumCPUsPerWorker   int32             `yaml:"num_cpus_per_worker"`
	PlacementStrategy  string            `yaml:"placement_strategy,omitempty"`
	ResourcesPerWorker map[string]string `yaml:"resources_per_worker,omitempty"`
}

type EngineConfig struct {
	ModelID         string                 `yaml:"model_id"`
	HFModelID       string                 `yaml:"hf_model_id,omitempty"`
	S3MirrorConfig  MirrorConfig           `yaml:"s3_mirror_config,omitempty"`
	GCSMirrorConfig MirrorConfig           `yaml:"gcs_mirror_config,omitempty"`
	Type            string                 `yaml:"type"`
	EngineKwargs    map[string]interface{} `yaml:"engine_kwargs"`
	MaxTotalTokens  int32                  `yaml:"max_total_tokens"`
	Generation      GenerationConfig       `yaml:"generation"`
}

type MirrorConfig struct {
	BucketURI string `yaml:"bucket_uri,omitempty"`
}

type GenerationConfig struct {
	PromptFormat      PromptFormat `yaml:"prompt_format"`
	StoppingSequences []string     `yaml:"stopping_sequences"`
}

type PromptFormat struct {
	System                            string `yaml:"system"`
	Assistant                         string `yaml:"assistant"`
	TrailingAssistant                 string `yaml:"trailing_assistant"`
	User                              string `yaml:"user"`
	DefaultSystemMessage              string `yaml:"default_system_message"`
	SystemInUser                      bool   `yaml:"system_in_user"`
	AddSystemTagsEvenIfMessageIsEmpty bool   `yaml:"add_system_tags_even_if_message_is_empty"`
	StripWhitespace                   bool   `yaml:"strip_whitespace"`
}

type DeploymentConfig struct {
	AutoScalingConfig    AutoScalingConfig `yaml:"auto_scaling_config"`
	MaxConcurrentQueries int32             `yaml:"max_concurrent_queries"`
	RayActorOptions      RayActorOptions   `yaml:"ray_actor_options"`
}

type RayActorOptions struct {
	Resources map[string]string `yaml:"resources,omitempty"`
}

type AutoScalingConfig struct {
	MinReplicas                        int32   `yaml:"min_replicas"`
	MaxReplicas                        int32   `yaml:"max_replicas"`
	InitialReplicas                    int32   `yaml:"initial_replicas"`
	TargetNumOngoingRequestsPerReplica int32   `yaml:"target_num_ongoing_requests_per_replica"`
	MetricsIntervalS                   float32 `yaml:"metrics_interval_s"`
	LookBackPeriodS                    float32 `yaml:"look_back_period_s"`
	SmoothingFactor                    float32 `yaml:"smoothing_factor"`
	DownscaleDelayS                    float32 `yaml:"downscale_delay_s"`
	UpscaleDelayS                      float32 `yaml:"upscale_delay_s"`
}

const (
	MaxConcurrentRatio = 40
)

func generateRayLLMModelConfig(modelTmpVersion *mlv1.ModelTemplateVersion) (string, error) {
	rayLLMModelConfig := RayLLMModelConfig{}
	rayLLMModelConfig.DeploymentConfig = setDeploymentConfig(modelTmpVersion)
	rayLLMModelConfig.ScalingConfig = setScalingConfig(modelTmpVersion)

	engineConfig, err := setEngineConfig(modelTmpVersion)
	if err != nil {
		return "", err
	}
	rayLLMModelConfig.EngineConfig = engineConfig

	yamlModelConfig, err := yaml.Marshal(&rayLLMModelConfig)
	if err != nil {
		return "", fmt.Errorf("failed to convert to YAML config, error: %s", err.Error())

	}
	return string(yamlModelConfig), nil
}

func setEngineConfig(model *mlv1.ModelTemplateVersion) (EngineConfig, error) {
	modelEngineConfig := model.Spec.EngineConfig
	prompt := modelEngineConfig.Generation.PromptFormat

	engineConfig := EngineConfig{
		ModelID:        model.Spec.ModelID,
		Type:           string(model.Spec.EngineConfig.Type),
		MaxTotalTokens: modelEngineConfig.MaxTotalTokens,
		Generation: GenerationConfig{
			PromptFormat: PromptFormat{
				System:                            prompt.System,
				Assistant:                         prompt.Assistant,
				TrailingAssistant:                 prompt.TrailingAssistant,
				User:                              prompt.User,
				DefaultSystemMessage:              prompt.DefaultSystemMessage,
				SystemInUser:                      prompt.SystemInUser,
				AddSystemTagsEvenIfMessageIsEmpty: prompt.AddSystemTagsEvenIfMessageIsEmpty,
				StripWhitespace:                   prompt.StripWhitespace,
			},
			StoppingSequences: modelEngineConfig.Generation.StoppingSequences,
		},
	}

	if modelEngineConfig.VLLMArgs != "" {
		if err := yaml.Unmarshal([]byte(modelEngineConfig.VLLMArgs), &engineConfig.EngineKwargs); err != nil {
			return engineConfig, fmt.Errorf("failed to convert vllmArgs, error: %s", err.Error())
		}
	} else {
		engineConfig.EngineKwargs = map[string]interface{}{
			"trust_remote_code":      true,
			"max_num_batched_tokens": engineConfig.MaxTotalTokens,
			"max_num_seq":            32,
			"gpu_memory_utilization": 0.9,
		}
	}

	// set default generation config if its value is empty string
	engineConfig.Generation = setDefaultGeneration(engineConfig.Generation)

	if err := setPrivateModel(&engineConfig, model); err != nil {
		return engineConfig, err
	}

	return engineConfig, nil
}

func setDefaultGeneration(generation GenerationConfig) GenerationConfig {
	promptFormat := generation.PromptFormat
	if promptFormat.System == "" {
		promptFormat.System = "<<SYS>>\\n{instruction}\\n<</SYS>>\\n\\n"
	}
	if promptFormat.Assistant == "" {
		promptFormat.Assistant = " {instruction} </s><s>"
	}
	if promptFormat.User == "" {
		promptFormat.User = "[INST] {system}{instruction} [/INST]"
	}
	if promptFormat.SystemInUser == false {
		promptFormat.SystemInUser = true
	}
	generation.PromptFormat = promptFormat
	if generation.StoppingSequences == nil {
		generation.StoppingSequences = []string{"\"<unk>\""}
	}
	return generation
}

func setScalingConfig(model *mlv1.ModelTemplateVersion) ScalingConfig {
	modelScalingConfig := model.Spec.ScalingConfig
	return ScalingConfig{
		NumWorkers:         modelScalingConfig.NumWorkers,
		NumGPUsPerWorker:   1,
		NumCPUsPerWorker:   modelScalingConfig.NumCPUsPerWorker,
		PlacementStrategy:  string(modelScalingConfig.PlacementStrategy),
		ResourcesPerWorker: modelScalingConfig.ResourcesPerWorker,
	}
}

func setDeploymentConfig(model *mlv1.ModelTemplateVersion) DeploymentConfig {
	modelDeploymentConfig := model.Spec.DeploymentConfig
	maxConcurrentQueries := modelDeploymentConfig.MaxConcurrentQueries
	if modelDeploymentConfig.TargetNumOngoingRequests == 0 {
		modelDeploymentConfig.TargetNumOngoingRequests = (maxConcurrentQueries * MaxConcurrentRatio) / 100
	}
	deploymentConfig := DeploymentConfig{
		AutoScalingConfig: AutoScalingConfig{
			MinReplicas:                        modelDeploymentConfig.MinReplicas,
			MaxReplicas:                        modelDeploymentConfig.MaxReplicas,
			InitialReplicas:                    modelDeploymentConfig.Replicas,
			TargetNumOngoingRequestsPerReplica: modelDeploymentConfig.TargetNumOngoingRequests,
			MetricsIntervalS:                   10.0,
			LookBackPeriodS:                    30.0,
			SmoothingFactor:                    0.6,
			DownscaleDelayS:                    300.0,
			UpscaleDelayS:                      60.0,
		},
		MaxConcurrentQueries: maxConcurrentQueries,
	}
	return deploymentConfig
}

func setPrivateModel(engineConfig *EngineConfig, model *mlv1.ModelTemplateVersion) error {
	if model.Spec.MirrorConfig != "" {
		if strings.Contains(model.Spec.MirrorConfig, "s3://") {
			engineConfig.S3MirrorConfig.BucketURI = model.Spec.MirrorConfig
		} else if strings.Contains(model.Spec.MirrorConfig, "gs://") {
			engineConfig.GCSMirrorConfig.BucketURI = model.Spec.MirrorConfig
		} else {
			return fmt.Errorf("invalid mirror config: %s", model.Spec.MirrorConfig)
		}
	} else if engineConfig.HFModelID == "" {
		engineConfig.HFModelID = model.Spec.ModelID
	}
	return nil
}
