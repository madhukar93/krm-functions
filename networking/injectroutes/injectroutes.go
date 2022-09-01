package injectroutes

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	yml "sigs.k8s.io/yaml"

	cv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	traefik "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	kv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	kindNetworking       = "IngressRoute"
	apiVersionNetworking = "traefik.containo.us/v1alpha1"
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

type RouteConfig struct {
	Match string `yaml:"match" ,json:"match"`
	Vpn   bool   `yaml:"vpn" ,json:"vpn"`
}

type functionConfig struct {
	App    string        `yaml:"app" ,json:"app"`
	Hosts  []string      `yaml:"hosts" ,json:"hosts"`
	Routes []RouteConfig `yaml:"routes" ,json:"routes"`
}

// change route to our own object
// create inject routes file
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

	fn, err := unwrap(fnConfig)
	if err != nil {
		return items, err
	}

	// check for deployment with app label
	foundDeployment := true
	for _, item := range items {
		_, err := item.GetMeta()
		if err != nil {
			return items, err
		}

		if fn.App == deploymentName {
			foundDeployment = true
		}
	}

	if !foundDeployment {
		return items, fmt.Errorf("could not find deployment with name: %s", fn.App)
	}

	foundIngressRoute := false
	for _, item := range items {

		meta, err := item.GetMeta()
		if err != nil {
			return items, err
		}

		if meta.Kind == kindNetworking && meta.APIVersion == apiVersionNetworking {
			foundIngressRoute = true
			// routes, err := item.GetSlice("spec.routes")
			routes, err := item.Pipe(yaml.Lookup("spec", "routes"))
			if err != nil {
				return items, err
			}

			inputRoutes := fn.Routes // get all of the input routes from fn
			//tempRoutes := []traefik.Route{}
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

				tempRoute := traefik.Route{}
				for i, route := range rts {
					if route.Match == exp {
						tempRoute.Match = exp
						tempRoute.Kind = "Rule"

						// service
						service := traefik.Service{}
						service.LoadBalancerSpec.Name = deploymentName
						service.LoadBalancerSpec.Port = intstr.FromInt(80)

						tempRoute.Services = append(tempRoute.Services, service)

						if inputRoute.Vpn {
							tempRoute.Middlewares = append(tempRoute.Middlewares, traefik.MiddlewareRef{
								Name:      "vpn-only",
								Namespace: "traefik",
							})
						}
						rts[i] = tempRoute
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
				tempRoute.Match = exp
				tempRoute.Kind = "Rule"
				// service
				service := traefik.Service{}
				service.LoadBalancerSpec.Name = deploymentName
				service.LoadBalancerSpec.Port = intstr.FromInt(80)

				tempRoute.Services = append(tempRoute.Services, service)

				if inputRoute.Vpn {
					tempRoute.Middlewares = append(tempRoute.Middlewares, traefik.MiddlewareRef{
						Name:      "vpn-only",
						Namespace: "traefik",
					})
				}

				rts = append(rts, tempRoute)
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

			serviceNode, err := generateService(fn, deploymentPort)
			if err != nil {
				return items, err
			}

			certificateNode, err := generateCertificate(fn)
			if err != nil {
				return items, err
			}

			items = append(items, serviceNode, certificateNode)
		}
	}

	if !foundIngressRoute {
		ingressRoute := traefik.IngressRoute{
			TypeMeta: metav1.TypeMeta{
				Kind:       kindNetworking,
				APIVersion: apiVersionNetworking,
			},

			ObjectMeta: metav1.ObjectMeta{
				Name: fn.App,
			},

			Spec: traefik.IngressRouteSpec{},
		}

		for _, inputRoute := range fn.Routes {
			hosts := makeCopy(fn.Hosts)

			exp, err := createMatchExpression(hosts, inputRoute.Match)
			if err != nil {
				return items, err
			}
			// service
			service := traefik.Service{}
			service.LoadBalancerSpec.Name = deploymentName
			service.LoadBalancerSpec.Port = intstr.FromInt(80)

			newRoute := traefik.Route{
				Match: exp,
				Kind:  "Rule",
				Services: []traefik.Service{
					service,
				},
			}

			if inputRoute.Vpn {
				newRoute.Middlewares = append(newRoute.Middlewares, traefik.MiddlewareRef{
					Name:      "vpn-only",
					Namespace: "traefik",
				})
			}
			ingressRoute.Spec.Routes = append(ingressRoute.Spec.Routes, newRoute)
		}

		ingressRouteNode, err := toRNode(ingressRoute)
		if err != nil {
			return items, err
		}

		serviceNode, err := generateService(fn, deploymentPort)
		if err != nil {
			return items, err
		}

		certificateNode, err := generateCertificate(fn)
		if err != nil {
			return items, err
		}

		items = append(items, ingressRouteNode, serviceNode, certificateNode)
	}
	return items, nil
}

func unwrap(fnConfig *yaml.RNode) (*functionConfig, error) {
	fn := &functionConfig{}
	// unmarshall fnConfig into struct
	data, err := fnConfig.GetFieldValue("data")
	if err != nil {
		return nil, err
	}

	fnYml, err := yml.Marshal(data)
	if err != nil {
		return nil, err
	}

	if err := yml.Unmarshal(fnYml, &fn); err != nil {
		return nil, err
	}

	return fn, nil
}

func toRNode(obj interface{}) (*yaml.RNode, error) {
	switch v := obj.(type) {
	case cv1.Certificate:
		{
			j, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}

			node, err := yaml.ConvertJSONToYamlNode(string(j))
			if err != nil {
				return nil, err
			}
			return node, nil
		}

	case kv1.Service:
		{
			j, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}

			node, err := yaml.ConvertJSONToYamlNode(string(j))
			if err != nil {
				return nil, err
			}
			return node, nil
		}

	case traefik.IngressRoute:
		{
			j, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}

			node, err := yaml.ConvertJSONToYamlNode(string(j))
			if err != nil {
				return nil, err
			}
			return node, nil
		}

	default:
		{
			fmt.Println(v)
			return nil, errors.New("unknown type can't convert")
		}
	}
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

func generateService(fn *functionConfig, deploymentPort int32) (*yaml.RNode, error) {
	// create a service object over here
	service := kv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fn.App,
		},
		Spec: kv1.ServiceSpec{
			Ports: []kv1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.IntOrString(intstr.FromInt(int(deploymentPort))),
				},
			},
			Selector: map[string]string{
				"app": fn.App,
			},
		},
	}

	// append service to the file
	serviceNode, err := toRNode(service)
	if err != nil {
		return nil, err
	}
	return serviceNode, nil
}

func generateCertificate(fn *functionConfig) (*yaml.RNode, error) {

	certificate := cv1.Certificate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Certificate",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fn.App,
		},
		Spec: cv1.CertificateSpec{
			SecretName: fn.App + "-cert",
			DNSNames:   fn.Hosts,
			IssuerRef: cmmeta.ObjectReference{
				Name: "letsencrypt",
				Kind: "ClusterIssuer",
			},
		},
	}

	certificateNode, err := toRNode(certificate)
	if err != nil {
		return nil, err
	}

	return certificateNode, nil
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