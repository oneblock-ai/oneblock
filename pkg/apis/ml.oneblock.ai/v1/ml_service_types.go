package v1

import (
	"github.com/rancher/wrangler/v2/pkg/condition"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/oneblock-ai/oneblock/pkg/apis/management.oneblock.ai/v1"
)

var (
	MLServiceCreated condition.Cond = "created"
	MLServiceReady   condition.Cond = "ready"
	MLServicePending condition.Cond = "pending"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:scope=Namespaced
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="ServiceStatus",type=string,JSONPath=".status.rayServiceStatuses.serviceStatus"

// MLService is the Schema for the LLM application service
type MLService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MLServiceSpec   `json:"spec,omitempty"`
	Status MLServiceStatus `json:"status,omitempty"`
}

type MLServiceSpec struct {
	// +kubebuilder:validation:Required
	ModelTemplateVersionRef *ModelTemplateVersionRef `json:"modelTemplateVersionRef"`
	// optional
	HFSecretRef *HFSecretRef `json:"hfSecretRef,omitempty"`
	// +kubebuilder:validation:Required
	MLClusterRef *MLClusterRef `json:"mlClusterRef"`
}

type HFSecretRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	SecretKey string `json:"secretKey,omitempty"`
}

type MLClusterRef struct {
	Name           string         `json:"name,omitempty"`
	Namespace      string         `json:"namespace,omitempty"`
	RayClusterSpec RayClusterSpec `json:"rayClusterSpec,omitempty"`
}

type RayClusterSpec struct {
	Image           string            `json:"image"`
	WorkerGroupSpec []WorkerGroupSpec `json:"workerGroupSpec,omitempty"`
}

type WorkerGroupSpec struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// Replicas is the number of desired Pods for this worker group. See https://github.com/ray-project/kuberay/pull/1443 for more details about the reason for making this field optional.
	// +kubebuilder:default:=1
	Replicas *int32 `json:"replicas,omitempty"`
	// MinReplicas denotes the minimum number of desired Pods for this worker group.
	// +kubebuilder:default:=1
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// MaxReplicas denotes the maximum number of desired Pods for this worker group, and the default value is maxInt32.
	// +kubebuilder:default:=5
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`
	// RayStartParams are the params of the start command: address, object-store-memory, ...
	RayStartParams   map[string]string `json:"rayStartParams,omitempty"`
	AcceleratorTypes map[string]int    `json:"acceleratorTypes,omitempty"`
	// Template is a pod template for the worker
	//Template corev1.PodTemplateSpec `json:"template"`
	Resources    *corev1.ResourceRequirements `json:"resources,omitempty"`
	NodeSelector map[string]string            `json:"nodeSelector,omitempty"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations      []corev1.Toleration `json:"tolerations,omitempty"`
	RuntimeClassName *string             `json:"runtimeClassName,omitempty"`
	Volume           *Volume             `json:"volume,omitempty"`
}

type ModelTemplateVersionRef struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
}

type MLServiceStatus struct {
	// Conditions is an array of current conditions
	Conditions         []v1.Condition           `json:"conditions,omitempty"`
	RayServiceStatuses rayv1.RayServiceStatuses `json:"rayServiceStatuses,omitempty"`
}
