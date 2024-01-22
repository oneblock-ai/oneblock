package main

import (
	"os"

	nvidiav1 "github.com/NVIDIA/gpu-operator/api/v1"
	controllergen "github.com/rancher/wrangler/v2/pkg/controller-gen"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/args"
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	volschv1 "volcano.sh/apis/pkg/apis/scheduling/v1beta1"
)

const (
	oneblockCoreGV      = "ml.oneblock.ai"
	oneblockMgmtGV      = "management.oneblock.ai"
	kubeRayGV           = "ray.io"
	nvidiaGV            = "nvidia.com"
	volcanoSchedulingGV = "scheduling.volcano.sh"
)

func main() {
	os.Unsetenv("GOPATH")
	controllergen.Run(args.Options{
		OutputPackage: "github.com/oneblock-ai/oneblock/pkg/generated",
		Boilerplate:   "hack/boilerplate.go.txt",
		Groups: map[string]args.Group{
			oneblockCoreGV: {
				PackageName: oneblockCoreGV,
				Types: []interface{}{
					// All structs with an embedded ObjectMeta field will be picked up
					"./pkg/apis/ml.oneblock.ai/v1",
				},
				GenerateTypes:   true,
				GenerateClients: true,
			},
			oneblockMgmtGV: {
				PackageName: oneblockMgmtGV,
				Types: []interface{}{
					// All structs with an embedded ObjectMeta field will be picked up
					"./pkg/apis/management.oneblock.ai/v1",
				},
				GenerateTypes:   true,
				GenerateClients: true,
			},
			kubeRayGV: {
				PackageName: kubeRayGV,
				Types: []interface{}{
					rayv1.RayCluster{},
					rayv1.RayJob{},
					rayv1.RayService{},
				},
				GenerateTypes:   false,
				GenerateClients: true,
			},
			nvidiaGV: {
				PackageName: nvidiaGV,
				Types: []interface{}{
					nvidiav1.ClusterPolicy{},
				},
				GenerateTypes:   false,
				GenerateClients: true,
			},
			volcanoSchedulingGV: {
				PackageName: volcanoSchedulingGV,
				Types: []interface{}{
					volschv1.PodGroup{},
					volschv1.Queue{},
				},
				GenerateTypes:   false,
				GenerateClients: true,
			},
		},
	})
}
