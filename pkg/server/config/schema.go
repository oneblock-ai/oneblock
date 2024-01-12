package config

import (
	"github.com/rancher/wrangler/v2/pkg/schemes"
	kuberayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

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
}
