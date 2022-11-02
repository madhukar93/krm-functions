package main

import (
	"fmt"
	"os"

	workloads "github.com/bukukasio/krm-functions/pkg/workloads"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func cmd() *cobra.Command {

	config := workloads.JobFunctionConfig{}
	p := framework.SimpleProcessor{
		Filter: kio.FilterFunc(config.Filter),
		Config: &config,
	}
	cmd := command.Build(p, command.StandaloneEnabled, false)
	cmd.Short = ""
	cmd.Long = ""
	return cmd
}

func main() {
	if err := cmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
