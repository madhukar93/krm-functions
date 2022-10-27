package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

/*
TODO
- preserve comments
- create generate/transform semantics in the framework
- try to leverage more of kyaml/fn package
- have validations
*/

func cmd() *cobra.Command {

	config := functionConfig{}
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
