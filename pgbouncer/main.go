package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

func cmd() *cobra.Command {
	cmd := command.Build(framework.ResourceListProcessorFunc(Process), command.StandaloneEnabled, false)
	cmd.Short = "generate pgbouncer resources for function config"
	cmd.Long = `
	This function generates pgbouncer resources for function config -
	deployment with pgbouncer container, and it's prometheus exporter sidecar
	service for the deployment
	pod disruption budget for the deployment to ensure min of 1 instance for a service is running
	pod monitor to collect prometheus metrics
	`
	return cmd
}

func main() {
	if err := cmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// TODO: use generics and put this in fnutils/ find a framework native way to do this
// Process filters the input resource list using the function config
func Process(resourceList *framework.ResourceList) error {
	if resourceList.FunctionConfig == nil {
		return fmt.Errorf("function config not found in resource list")
	}
	fnConfig := &functionConfig{}

	// TODO: use openapi spec to validate function config
	// the below does schema validation on the function config as well
	// and runs custom validators
	err := framework.LoadFunctionConfig(resourceList.FunctionConfig, fnConfig)
	if err != nil {
		return err
	}
	resourceList.Items, _ = fnConfig.filter(resourceList.Items)
	return nil
}
