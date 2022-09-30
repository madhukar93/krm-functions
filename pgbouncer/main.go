package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

func main() {
	cmd := command.Build(framework.ResourceListProcessorFunc(filter), command.StandaloneEnabled, false)
	cmd.Short = "Inject files wrapped in KRM resources into ConfigMap keys"
	cmd.Long = "Inject files or templates wrapped in KRM resources into ConfigMap keys"

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func filter(resourceList *framework.ResourceList) error {
	if resourceList.FunctionConfig == nil {
		return fmt.Errorf("no function config specified")
	}
	fnConfig := &functionConfig{}
	err := framework.LoadFunctionConfig(resourceList.FunctionConfig, fnConfig)
	if err != nil {
		return err
	}
	output, err := fnConfig.filter(resourceList.Items)
	if err != nil {
		resourceList.Results = framework.Results{
			&framework.Result{
				Message:  err.Error(),
				Severity: framework.Error,
			},
		}
		return resourceList.Results
	}
	resourceList.Items = output
	// results, err := injector.Results()
	// if err != nil {
	// 	resourceList.Results = framework.Results{
	// 		&framework.Result{
	// 			Message:  err.Error(),
	// 			Severity: framework.Error,
	// 		},
	// 	}
	// 	return resourceList.Results
	// }
	// resourceList.Results = results
	return nil
}
