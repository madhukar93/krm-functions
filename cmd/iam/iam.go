package main

import (
	"fmt"
	"os"

	iam "github.com/bukukasio/krm-functions/pkg/iam"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

func cmd() *cobra.Command {
	p := &framework.VersionedAPIProcessor{FilterProvider: framework.GVKFilterMap{
		"LummoIAM": {
			"krm/v1": &iam.LummoIAM{},
		},
	},
	}
	cmd := command.Build(p, command.StandaloneEnabled, false)
	cmd.Short = ""
	cmd.Long = "This function generates GSA, KSA and IAM binding policies for both SA's in GCP project and Kubernetes."
	return cmd
}

func main() {
	if err := cmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
