//go:generate go run pkg/codegen/main.go
//go:generate /bin/rm -rf pkg/generated
//go:generate go run pkg/codegen/main.go
//go:generate /bin/bash hack/generate-manifest
package main

import (
	"os"

	"github.com/rancher/wrangler/v2/pkg/signals"
	"github.com/sirupsen/logrus"

	"github.com/oneblock-ai/oneblock/cmd"
)

func main() {
	cmd := cmd.New()

	ctx := signals.SetupSignalContext()
	if err := cmd.ExecuteContext(ctx); err != nil {
		logrus.Fatal(err.Error())
		os.Exit(1)
	}
}
