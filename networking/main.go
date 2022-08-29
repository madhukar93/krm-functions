package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/bukukasio/kpt-functions/inject-routes/injectroutes"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

func main() {
	// -- uncomment below lines and run kpt fn source data | go run main.go to check function output -- //
	// file, _ := os.Open("./data/fn.yaml")
	// defer file.Close()
	// os.Stdin = file

	p := InjectRouteProcessor{}
	cmd := command.Build(&p, command.StandaloneEnabled, false)

	cmd.Short = "Inject files wrapped in KRM resources into ConfigMap keys"
	cmd.Long = "Inject files or templates wrapped in KRM resources into ConfigMap keys"

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type InjectRouteProcessor struct{}

func (p *InjectRouteProcessor) Process(resourceList *framework.ResourceList) error {
	fnConfig := resourceList.FunctionConfig

	if fnConfig == nil {
		return errors.New("no function config specified")
	}
	injector, err := injectroutes.New(fnConfig)
	if err != nil {
		return err
	}
	items, err := injector.Filter(resourceList.Items)
	if err != nil {
		resourceList.Results = framework.Results{
			&framework.Result{
				Message:  err.Error(),
				Severity: framework.Error,
			},
		}
		return resourceList.Results
	}
	resourceList.Items = items

	results, err := injector.Results()
	if err != nil {
		resourceList.Results = framework.Results{
			&framework.Result{
				Message:  err.Error(),
				Severity: framework.Error,
			},
		}
		return resourceList.Results
	}
	resourceList.Results = results

	return nil
}
