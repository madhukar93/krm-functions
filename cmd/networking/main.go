package main

import (
	"errors"
	"fmt"
	"os"

	networking "github.com/bukukasio/krm-functions/pkg/networking"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

func main() {
	cmd := command.Build(framework.ResourceListProcessorFunc(Process), command.StandaloneEnabled, false)

	cmd.Short = "Inject files wrapped in KRM resources into ConfigMap keys"
	cmd.Long = "Inject files or templates wrapped in KRM resources into ConfigMap keys"

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Process(resourceList *framework.ResourceList) error {
	fnConfig := resourceList.FunctionConfig

	if fnConfig == nil {
		return errors.New("no function config specified")
	}
	injector, err := networking.FnConfigFromRNode(fnConfig)
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
