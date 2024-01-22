package config

import (
	nvidiav1 "github.com/NVIDIA/gpu-operator/api/v1"
	"github.com/rancher/wrangler/v2/pkg/schemes"
	kuberayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	schedulingv1beta1 "volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	oneblockaiv1 "github.com/oneblock-ai/oneblock/pkg/apis/ml.oneblock.ai/v1"
)

var (
	localSchemeBuilder = runtime.SchemeBuilder{
		oneblockaiv1.AddToScheme,
		kuberayv1.AddToScheme,
	}
	AddToScheme = localSchemeBuilder.AddToScheme
	Scheme      = runtime.NewScheme()
)

func init() {
	utilruntime.Must(AddToScheme(Scheme))
	utilruntime.Must(schemes.AddToScheme(Scheme))
	utilruntime.Must(oneblockaiv1.AddToScheme(Scheme))
	utilruntime.Must(kuberayv1.AddToScheme(Scheme))
	utilruntime.Must(nvidiav1.AddToScheme(Scheme))
	utilruntime.Must(schedulingv1beta1.AddToScheme(Scheme))
}
