package v1

import (
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/oneblock-ai/oneblock/pkg/apis/management.oneblock.ai/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:scope=Namespaced
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=`.metadata.creationTimestamp`

// MLServe is the Schema for the LLM application service
type MLServe struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServeSpec   `json:"spec,omitempty"`
	Status ServeStatus `json:"status,omitempty"`
}

type ServeSpec struct {
	Applications []ServeApplication `json:"applications,omitempty"`
	MLClusterRef MLClusterRef       `json:"mlClusterRef,omitempty"`
	ModelRef     ModelTemplateRef   `json:"modelRef"`
}

type ModelTemplateRef struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	// use default model version if this is not specified
	Version string `json:"version,omitempty"`
}

type ServeApplication struct {
	Name          string               `json:"name,omitempty"`
	URL           string               `json:"url,omitempty"`
	ImportPath    string               `json:"importPath,omitempty"`
	RoutePrefix   string               `json:"routePrefix,omitempty"`
	LoggingConfig ServiceLoggingConfig `json:"loggingConfig,omitempty"`
	Deployments   []ServiceDeployment  `json:"deployments,omitempty"`
	RuntimeEnv    map[string][]string  `json:"runtimeEnv,omitempty"`
	Args          map[string][]string  `json:"args,omitempty"`
}

type ServiceLoggingConfig struct {
	EnableAccessLog bool   `json:"enableAccessLog,omitempty"`
	Encoding        string `json:"encoding,omitempty"` // default to 'TEXT'. 'JSON' is also supported to format all serve logs into json structure.
	LogLevel        int    `json:"logLevel,omitempty"`
	LogDir          string `json:"logDir,omitempty"`
}

type ServiceDeployment struct {
	Name                      string   `json:"name,omitempty"`
	NumReplicas               int32    `json:"numReplicas,omitempty"`
	MaxConcurrentQueries      int32    `json:"maxConcurrentQueries,omitempty"`
	GracefulShutdownWaitLoopS string   `json:"gracefulShutdownWaitLoopS,omitempty"`
	GracefulShutdownTimeoutS  string   `json:"gracefulShutdownTimeoutS,omitempty"`
	HealthCheckPeriodS        string   `json:"healthCheckPeriodS,omitempty"`
	HealthCheckTimeoutS       string   `json:"healthCheckTimeoutS,omitempty"`
	RayActorOptions           []string `json:"rayActorOptions,omitempty"`
	UserConfig                []string `json:"userConfig,omitempty"`
}

type MLClusterRef struct {
	Name          string               `json:"name,omitempty"`
	Namespace     string               `json:"namespace,omitempty"` // default to the same serve namespace if not specified
	ClusterConfig rayv1.RayClusterSpec `json:"clusterConfig,omitempty"`
}

type ServeStatus struct {
	// Conditions is an array of current conditions
	Conditions []v1.Condition        `json:"conditions"`
	State      corev1.ContainerState `json:"state"`
}
