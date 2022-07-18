package injectroutes

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type injectResult struct {
	Source   *yaml.RNode
	Target   *yaml.RNode
	Keys     []string
	ErrorMsg string
}

type InjectRoutes struct {
	injectResults []*injectResult
}

func (i *InjectRoutes) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, item := range items {
		meta, err := item.GetMeta()
		if err != nil {
			return nil, err
		}

		if meta.Kind == "IngressRoute" && meta.APIVersion == "traefik.containo.us/v1alpha1" {
			routes, err := item.GetSlice("spec.routes")
			if err != nil {
				return nil, err
			}

			for _, inputRoute := range functionConfig.Data.Routes {
				for _, route := range routes {
					exp, err := createMatchExpression(functionConfig.Domains, inputRoute)
					if err != nil {
						return nil, err
					}
					if route == exp {
						return items, nil
					}

					// create new route object from function config and append to routes
				}
			}
		}
	}
	return items, nil
}

func (i *InjectRoutes) Results() (framework.Results, error) {
	var results framework.Results
	if len(i.injectResults) == 0 {
		results = append(results, &framework.Result{
			Message: "no injections",
		})
		return results, nil
	}
	for _, injectResult := range i.injectResults {
		fmt.Println(injectResult)
	}
	return results, nil
}

func createMatchExpression(domains []string, expression string) (string, error) {
	if expression == "" {
		return "", fmt.Errorf("input string is empty")
	}
	for i, domain := range domains {
		domains[i] = fmt.Sprintf("Host(`%s`)", domain)
	}
	newExpression := strings.Join(domains, " || ")
	return newExpression, nil
}
