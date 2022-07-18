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

func (i *InjectRoutes) Filter(items []*yaml.RNode, fnConfig *yaml.RNode) ([]*yaml.RNode, error) {
	for _, item := range items {
		meta, err := item.GetMeta()
		if err != nil {
			return nil, err
		}

		if meta.Kind == "IngressRoute" && meta.APIVersion == "traefik.containo.us/v1alpha1" {
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
			fmt.Println(fn.Route)
			////////////////////////////////////

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

			fmt.Println(rts)
			rtYaml, err = yaml.Marshal(rts)
			if err != nil {
				return items, err
			}
			routesObj, err := yaml.Parse(string(rtYaml))
			if err != nil {
				return items, err
			}

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
		fmt.Println(injectResult)
	}
	return results, nil
}

// func (i *InjectRoutes) injectRoutes(source *yaml.RNode, route *yaml.RNode) (*yaml.RNode, error) {
// 	// result - newinjectresul

// 	data, err := source.GetFieldValue("spec.routes")
// 	if err != nil {
// 		return route, err
// 	}

// 	dataMap, ok := data.([]map[string]interface{})
// 	if !ok {
// 		err = fmt.Errorf(
// 			"data must be a []map[string]interface, got %T",
// 			data,
// 		)
// 		// result.ErrorMsg = err.Error()
// 		return route, err
// 	}
// }

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
