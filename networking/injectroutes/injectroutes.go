package injectroutes

import (
	"errors"
	"fmt"
	"strings"

	yml "sigs.k8s.io/yaml"

	"github.com/bukukasio/kpt-functions/inject-routes/utils"
	traefik "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	kv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	App    string          `yaml:"app" ,json:"app"`
	Hosts  []string        `yaml:"hosts" ,json:"hosts"`
	Routes []traefik.Route `yaml:"routes" ,json:"routes"`
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

func (in *InjectRoutes) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	fnConfig := in.fnConfig
	result := &injectResult{} // this is optional, mainly for debugging and observability purposes

	// get the deploymeny information to generate services and certificates
	deploymentName, deploymentPort, err := getDeploymentData(items)

	if err != nil {
		return items, err
	}

	for _, item := range items {
		meta, err := item.GetMeta()
		if err != nil {
			return items, err
		}

		if meta.Kind == kind && meta.APIVersion == apiVersion {
			// routes, err := item.GetSlice("spec.routes")
			routes, err := item.Pipe(yaml.LookupCreate(yaml.SequenceNode, "spec", "routes"))
			if err != nil {
				return items, err
			}

			// unmarshall fnConfig into struct
			data, err := fnConfig.GetFieldValue("data")
			if err != nil {
				return items, err
			}

			// TODO: extra fields those not in the schema is ignored, we want to exit with error
			var fn functionConfig
			fnYml, err := yml.Marshal(data)
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
				hosts := makeCopy(fn.Hosts)

				exp, err := createMatchExpression(hosts, inputRoute.Match)
				if err != nil {
					return items, err
				}

				for i, route := range rts {
					if route.Match == exp {
						inputRoute.Match = exp
						inputRoute.Kind = "Rule"

						// service
						service := traefik.Service{}
						service.LoadBalancerSpec.Name = deploymentName
						service.LoadBalancerSpec.Port = intstr.FromInt(int(deploymentPort))

						inputRoute.Services = append(inputRoute.Services, service)

						rts[i] = inputRoute
						routesObj, err := setRoutes(rts)

						if err != nil {
							return items, err
						}

						err = item.PipeE(
							yaml.LookupCreate(yaml.MappingNode, "spec"),
							yaml.SetField("routes", routesObj))
						if err != nil {
							return items, err
						}

						result.Source = item
						result.Route = routesObj
						in.injectResults = append(in.injectResults, result)

						return items, nil
					}
				}
				inputRoute.Match = exp
				inputRoute.Kind = "Rule"
				// service
				service := traefik.Service{}
				service.LoadBalancerSpec.Name = deploymentName
				service.LoadBalancerSpec.Port = intstr.FromInt(int(deploymentPort))

				inputRoute.Services = append(inputRoute.Services, service)

				rts = append(rts, inputRoute)
			}

			routesObj, err := setRoutes(rts)
			if err != nil {
				return items, err
			}

			err = item.PipeE(
				yaml.LookupCreate(yaml.MappingNode, "spec"),
				yaml.SetField("routes", routesObj))
			if err != nil {
				return items, err
			}

			result.Source = item
			result.Route = routesObj
			in.injectResults = append(in.injectResults, result)

			// set app name
			err = item.PipeE(
				yaml.Lookup("metadata"),
				yaml.SetField("name", yaml.NewScalarRNode(fn.App)),
			)
			if err != nil {
				return items, err
			}
			err = item.PipeE(
				yaml.Lookup("spec", "tls"),
				yaml.SetField("secretName", yaml.NewScalarRNode(fn.App+"-cert")),
			)

			if err != nil {
				return items, err
			}

			// create a service object over here
			svc := kv1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: fn.App,
				},
				Spec: kv1.ServiceSpec{
					Ports: []kv1.ServicePort{
						{
							Port: int32(deploymentPort),
						},
					},
					Selector: map[string]string{
						"app": fn.App,
					},
				},
			}

			err = utils.CreateService("service.yaml", svc)
			if err != nil {
				return items, nil
			}
		}
	}

	return items, nil
}

func setRoutes(routes []traefik.Route) (*yaml.RNode, error) {
	// struct to yaml.RNode
	rtYaml, err := yml.Marshal(routes)
	if err != nil {
		return nil, err
	}

	obj, err := yaml.Parse(string(rtYaml))
	if err != nil {
		return nil, err
	}
	return obj, nil
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
			msg = fmt.Sprintf("injected route to source: %s", source)
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

func getDeploymentData(items []*yaml.RNode) (string, int32, error) {
	deployment := v1.Deployment{}
	for _, item := range items {
		meta, err := item.GetMeta()
		if err != nil {
			return "", 0, err
		}
		if meta.Kind == "Deployment" && meta.APIVersion == "apps/v1" {
			dYaml, err := yml.Marshal(item)
			if err != nil {
				return "", 0, err
			}

			if err := yml.Unmarshal(dYaml, &deployment); err != nil {
				return "", 0, err
			}
		}
	}
	return deployment.Spec.Template.Spec.Containers[0].Name, deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort, nil
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

func makeCopy(hosts []string) []string {
	tmp := make([]string, len(hosts))
	copy(tmp, hosts)
	return tmp
}
