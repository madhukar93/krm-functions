package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func cmd() *cobra.Command {
	config := functionConfig{}
	p := framework.SimpleProcessor{
		Filter: kio.FilterFunc(config.Filter),
		Config: &config,
	}
	cmd := command.Build(p, command.StandaloneEnabled, false)
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
