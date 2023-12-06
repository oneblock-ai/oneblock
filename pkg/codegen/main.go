package main

import (
	"os"

	controllergen "github.com/rancher/wrangler/v2/pkg/controller-gen"
	"github.com/rancher/wrangler/v2/pkg/controller-gen/args"
)

const (
	oneblockCoreGV = "core.oneblock.ai"
	oneblockMgmtGV = "management.oneblock.ai"
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
					"./pkg/apis/core.oneblock.ai/v1",
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
		},
	})
}
