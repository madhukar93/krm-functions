// package main

// import (
// 	"fmt"
// 	"io"
// 	"os"
// 	"strings"

// 	"sigs.k8s.io/kustomize/kyaml/kio"
// 	"sigs.k8s.io/kustomize/kyaml/yaml"
// )

// type filter struct {
// 	readWriter *kio.ByteReadWriter
// }

// type Route struct {
// 	Match       string       `yaml:"match"`
// 	Priority    int          `yaml:"priority,omitempty"`
// 	MiddleWares []Middleware `yaml:"middlewares,omitempty"`
// 	Services    []Service    `yaml:"services,omitempty"`
// }

// type Middleware struct {
// 	Name      string `yaml:"name"`
// 	Namespace string `yaml:"namespace"`
// }

// type Service struct {
// 	Name               string `yaml:"name,omitempty"`
// 	Namespace          string `yaml:"namespace,omitempty"`
// 	PassHostHeader     bool   `yaml:"passHostHeader,omitempty"`
// 	Port               int    `yaml:"port,omitempty"`
// 	ResponseForwarding struct {
// 		FlushInterval string `yaml:"flushInterval,omitempty"`
// 	} `yaml:"responseForwarding,omitempty"`
// 	Scheme           string `yaml:"scheme,omitempty"`
// 	ServersTransport string `yaml:"serversTransport,omitempty"`
// 	Sticky           struct {
// 		Cookie struct {
// 			HttpOnly bool   `yaml:"httpOnly,omitempty"`
// 			Name     string `yaml:"name,omitempty"`
// 			Secure   bool   `yaml:"secure,omitempty"`
// 			SameSite string `yaml:"sameSite,omitempty"`
// 		} `yaml:"cookie,omitempty"`
// 	} `yaml:"sticky,omitempty"`
// 	Strategy string `yaml:"strategy,omitempty"`
// 	Weight   int    `yaml:"weight,omitempty"`
// }
// type functionConfig struct {
// 	Data struct {
// 		Routes []Route `yaml:"routes"`
// 	} `yaml:"data"`
// }

// func main() {
// 	if err := RunPipeline(os.Stdin, os.Stdout); err != nil {
// 		fmt.Fprint(os.Stderr, err)
// 		os.Exit(1)
// 	}
// }

// func RunPipeline(reader io.Reader, writer io.Writer) error {
// 	readWriter := &kio.ByteReadWriter{
// 		Reader: reader,
// 		Writer: writer,
// 	}
// 	pipeline := kio.Pipeline{
// 		Inputs:  []kio.Reader{readWriter},
// 		Filters: []kio.Filter{filter{readWriter: readWriter}},
// 		Outputs: []kio.Writer{readWriter},
// 	}

// 	return pipeline.Execute()
// }

// func (f filter) Filter(in []*yaml.RNode) ([]*yaml.RNode, error) {
// 	out := []*yaml.RNode{}

// 	marshalledFunctionConfig, err := f.readWriter.FunctionConfig.String()
// 	if err != nil {
// 		return out, err
// 	}

// 	var config functionConfig
// 	if err := yaml.Unmarshal([]byte(marshalledFunctionConfig), &config); err != nil {
// 		return out, err
// 	}

// 	fmt.Println(config)

// 	for _, resource := range in {
// 		routesNode, err := resource.Pipe(yaml.Lookup("spec", "routes"))
// 		if err != nil {
// 			return out, err
// 		}
// 		routesList, err := routesNode.Elements()
// 		if err != nil {
// 			return out, err
// 		}

// 		//for _, inputRoute := range inputRoutesList {
// 		for _, route := range routesList {
// 			expression, err := route.GetString("match")
// 			if err != nil {
// 				return out, err
// 			}

// 			fmt.Println(expression)
// 			// match expression to input function config

// 		}
// 		//}
// 	}

// 	return out, nil
// }

// func createMatchExpression(domains []string, expression string) (string, error) {
// 	if expression == "" {
// 		return "", fmt.Errorf("input string is empty")
// 	}
// 	for i, domain := range domains {
// 		domains[i] = fmt.Sprintf("Host(`%s`)", domain)
// 	}
// 	newExpression := strings.Join(domains, " || ")
// 	return newExpression, nil
// }

package main

import (
	"fmt"
	"os"

	"github.com/bukukasio/kpt-network-resource/injectroutes"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

func main() {
	file, _ := os.Open("./test/fn.yaml")
	defer file.Close()
	os.Stdin = file

	p := ConfigMapInjectorProcessor{}
	cmd := command.Build(&p, command.StandaloneEnabled, false)

	cmd.Short = "Inject files wrapped in KRM resources into ConfigMap keys"
	cmd.Long = "Inject files or templates wrapped in KRM resources into ConfigMap keys"

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type ConfigMapInjectorProcessor struct{}

func (p *ConfigMapInjectorProcessor) Process(resourceList *framework.ResourceList) error {
	injector := &injectroutes.InjectRoutes{}

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
