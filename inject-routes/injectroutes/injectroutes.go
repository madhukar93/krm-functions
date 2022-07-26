package injectroutes

import (
	"errors"
	"fmt"
	"strings"

	yml "github.com/ghodss/yaml"

	traefik "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	kind       = "IngressRoute"
	apiVersion = "traefik.containo.us/v1alpha1"
)

type injectResult struct {
	Source   *yaml.RNode
	Route    *yaml.RNode
	ErrorMsg string
}
type InjectRoutes struct {
	fnConfig      *yaml.RNode
	injectResults []*injectResult
}

type functionConfig struct {
	AppName string          `json:"app"`
	Hosts   []string        `json:"hosts"`
	Routes  []traefik.Route `json:"routes"`
}

func New(fnConfig *yaml.RNode) (*InjectRoutes, error) {
	if fnConfig == nil {
		return nil, errors.New("no functionConfig specified")
	}

	fn := &InjectRoutes{
		fnConfig: fnConfig,
	}
	return fn, nil
}

func (i *InjectRoutes) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	fnConfig := i.fnConfig

	result := &injectResult{} // this is optional, mainly for debugging and observability purposes

	for _, item := range items {
		meta, err := item.GetMeta()
		if err != nil {
			return items, err
		}

		if meta.Kind == kind && meta.APIVersion == apiVersion {
			routes, err := item.GetSlice("spec.routes")
			if err != nil {
				return items, err
			}

			// unmarshall fnConfig into struct
			inputRoute, err := fnConfig.GetFieldValue("data")
			if err != nil {
				return items, err
			}

			// TODO: extra fields those not in the schema is ignored, we want to exit with error
			var fn functionConfig
			fnYml, err := yml.Marshal(inputRoute)
			if err != nil {
				return items, err
			}

			if err := yml.Unmarshal(fnYml, &fn); err != nil {
				return items, err
			}

			inputRoutes := fn.Routes // get all of the input routes from fn

			// unmarshall routes into struct and perform operations
			rtYaml, err := yml.Marshal(routes)
			if err != nil {
				return items, err
			}
			var rts []traefik.Route
			if err := yml.Unmarshal(rtYaml, &rts); err != nil {
				return items, err
			}

			for _, inputRoute := range inputRoutes {
				exp, err := createMatchExpression(fn.Hosts, inputRoute.Match)
				if err != nil {
					return items, err
				}

				for _, route := range rts {
					if route.Match == exp {
						return items, nil
					}
				}

				inputRoute.Match = exp
				rts = append(rts, inputRoute)

				rtYaml, err = yml.Marshal(rts)
				if err != nil {
					return items, err
				}

				routesObj, err := yaml.Parse(string(rtYaml))
				if err != nil {
					return items, err
				}

				result.Source = item
				result.Route = routesObj
				i.injectResults = append(i.injectResults, result)

				err = item.PipeE(
					yaml.Lookup("spec"),
					yaml.SetField("routes", routesObj))
				if err != nil {
					return items, err
				}
			}

			// set app name
			err = item.PipeE(
				yaml.Lookup("metadata"),
				yaml.SetField("name", yaml.NewScalarRNode(fn.AppName)),
			)
			if err != nil {
				return items, err
			}
			err = item.PipeE(
				yaml.Lookup("spec", "tls"),
				yaml.SetField("secretName", yaml.NewScalarRNode(fn.AppName+"-cert")),
			)
			if err != nil {
				return items, err
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
		var (
			msg      string
			severity framework.Severity
			source   = fmt.Sprintf("%s %s", injectResult.Source.GetKind(), injectResult.Source.GetName())
			route    = injectResult.Route
		)
		if injectResult.ErrorMsg != "" {
			msg = fmt.Sprintf("%v failed to inject route to source %s: %s", route, source, injectResult.ErrorMsg)
			severity = framework.Error
		} else {
			msg = fmt.Sprintf("injected route: %v to source: %s", route, source)
			severity = framework.Info
		}

		result := &framework.Result{
			Message:  msg,
			Severity: severity,
		}

		results = append(results, result)
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
	newExpression = newExpression + fmt.Sprintf(" && %s", expression)
	return newExpression, nil
}
