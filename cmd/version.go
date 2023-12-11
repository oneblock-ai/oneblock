package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/oneblock-ai/oneblock/pkg/version"
)

func NewVersion(_ CommandContext) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of oneblock",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("oneblock version %s\n", version.FriendlyVersion())
		},
	}
}
