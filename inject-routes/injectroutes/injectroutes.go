package injectroutes

import (
	"errors"
	"fmt"
	"strings"

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

type Route struct {
	Kind        string       `yaml:"kind"`
	Match       string       `yaml:"match"`
	Priority    int          `yaml:"priority,omitempty"`
	MiddleWares []Middleware `yaml:"middlewares,omitempty"`
	Services    []Service    `yaml:"services,omitempty"`
}

type Middleware struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type Service struct {
	Name               string `yaml:"name,omitempty"`
	Namespace          string `yaml:"namespace,omitempty"`
	PassHostHeader     bool   `yaml:"passHostHeader,omitempty"`
	Port               int    `yaml:"port,omitempty"`
	ResponseForwarding struct {
		FlushInterval string `yaml:"flushInterval,omitempty"`
	} `yaml:"responseForwarding,omitempty"`
	Scheme           string `yaml:"scheme,omitempty"`
	ServersTransport string `yaml:"serversTransport,omitempty"`
	Sticky           struct {
		Cookie struct {
			HttpOnly bool   `yaml:"httpOnly,omitempty"`
			Name     string `yaml:"name,omitempty"`
			Secure   bool   `yaml:"secure,omitempty"`
			SameSite string `yaml:"sameSite,omitempty"`
		} `yaml:"cookie,omitempty"`
	} `yaml:"sticky,omitempty"`
	Strategy string `yaml:"strategy,omitempty"`
	Weight   int    `yaml:"weight,omitempty"`
}

type functionConfig struct {
	AppName string   `yaml:"app"`
	Hosts   []string `yaml:"hosts"`
	Route   Route    `yaml:"route"`
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

			var fn functionConfig
			fnYml, err := yaml.Marshal(inputRoute)
			if err != nil {
				return items, err
			}

			if err := yaml.Unmarshal(fnYml, &fn); err != nil {
				return items, err
			}

			// unmarshall routes into struct and perform operations
			rtYaml, err := yaml.Marshal(routes)
			if err != nil {
				return items, err
			}
			var rts []Route
			if err := yaml.Unmarshal(rtYaml, &rts); err != nil {
				return items, err
			}

			exp, err := createMatchExpression(fn.Hosts, fn.Route.Match)
			if err != nil {
				return items, err
			}

			for _, route := range rts {
				if route.Match == exp {
					return items, nil
				}
			}
			fn.Route.Match = exp
			rts = append(rts, fn.Route)

			rtYaml, err = yaml.Marshal(rts)
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

			_, err = item.Pipe(
				yaml.LookupCreate(yaml.MappingNode, "spec"),
				yaml.SetField("routes", routesObj))
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
