package networking

import (
	"errors"
	"fmt"
	"strings"

	yml "sigs.k8s.io/yaml"

	"github.com/bukukasio/krm-functions/pkg/common/fnutils"
	cv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	traefik "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	ingressRouteKind     = "IngressRoute"
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
	// FIXME: use yaml.NewYAMLOrJSONDecoder
	Match string `yaml:"match ,omitempty" ,json:"match, omitempty"`
	Vpn   bool   `yaml:"vpn ,omitempty" ,json:"vpn ,omitempty"`
}

type functionConfig struct {
	App    string        `yaml:"app" ,json:"app"`
	Hosts  []string      `yaml:"hosts" ,json:"hosts"`
	Grpc   bool          `yaml:"grpc ,omitempty" ,json:"grpc ,omitempty"`
	Routes []RouteConfig `yaml:"routes" ,json:"routes"`
}

// change route to our own object
// create inject routes file
func FnConfigFromRNode(fnConfig *yaml.RNode) (*InjectRoutes, error) {
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
	//result := &injectResult{} // this is optional, mainly for debugging and observability purposes

	// get the deploymeny information to generate services and certificates
	deploymentName, httpsPort, grpcPort, err := getDeploymentData(items)
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

	// delete all the existing services, ingress and certificates
	// FIXM: replace only the ones that this function creates
	out := []*yaml.RNode{}

	for _, resource := range items {
		meta, err := resource.GetMeta()
		if err != nil {
			return items, err
		}

		if meta.Kind != "Service" && meta.Kind != "IngressRoute" && meta.Kind != "Certificate" {
			out = append(out, resource)
		}
	}

	//if !fn.Grpc {
	ingressRoute := traefik.IngressRoute{
		TypeMeta: metav1.TypeMeta{
			Kind:       ingressRouteKind,
			APIVersion: apiVersionNetworking,
		},

		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-http", fn.App),
		},

		Spec: traefik.IngressRouteSpec{},
	}

	ingressRouteGrpc := traefik.IngressRoute{
		TypeMeta: metav1.TypeMeta{
			Kind:       ingressRouteKind,
			APIVersion: apiVersionNetworking,
		},

		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-grpc", fn.App),
		},

		Spec: traefik.IngressRouteSpec{},
	}

	for _, inputRoute := range fn.Routes {
		hosts := makeCopy(fn.Hosts)

		exp, err := createMatchExpression(hosts, inputRoute.Match)
		if err != nil {
			return nil, err
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

	if fn.Grpc {
		if grpcPort == 0 {
			// grpc port not found on deployment
			err = errors.New("grpc port not found on deployment")
			return nil, err
		}
		// service
		service := traefik.Service{}
		service.LoadBalancerSpec.Name = deploymentName
		service.LoadBalancerSpec.Port = intstr.FromInt(int(grpcPort))
		service.LoadBalancerSpec.PassHostHeader = &[]bool{true}[0] //TODO: some hack to create a pointer to bool
		service.LoadBalancerSpec.Scheme = "h2c"

		newRoute := traefik.Route{
			Match: fmt.Sprintf("Host(`%s.internal.bukukas.k8s`)", fn.App), // fake domain currently
			Kind:  "Rule",
			Services: []traefik.Service{
				service,
			},
		}
		ingressRouteGrpc.Spec.Routes = append(ingressRouteGrpc.Spec.Routes, newRoute)
	}

	ingressRouteNode, err := fnutils.MakeRNode(ingressRoute)
	if err != nil {
		return nil, err
	}

	ingressRouteNodeGrpc, err := fnutils.MakeRNode(ingressRouteGrpc)
	if err != nil {
		return nil, err
	}

	serviceNode, err := generateService(fn, httpsPort, grpcPort)
	if err != nil {
		return nil, err
	}

	certificateNode, err := generateCertificate(fn)
	if err != nil {
		return nil, err
	}

	if fn.Grpc {
		out = append(out, ingressRouteNode, ingressRouteNodeGrpc, serviceNode, certificateNode)
	} else {
		out = append(out, ingressRouteNode, serviceNode, certificateNode)
	}
	return out, nil
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

func getDeploymentData(items []*yaml.RNode) (string, int32, int32, error) {
	deployment := v1.Deployment{}
	for _, item := range items {
		meta, err := item.GetMeta()
		if err != nil {
			return "", 0, 0, err
		}
		if meta.Kind == "Deployment" && meta.APIVersion == "apps/v1" {
			dYaml, err := yml.Marshal(item)
			if err != nil {
				return "", 0, 0, err
			}

			if err := yml.Unmarshal(dYaml, &deployment); err != nil {
				return "", 0, 0, err
			}
		}
	}

	var httpsPort int32
	var grpcPort int32

	ports := deployment.Spec.Template.Spec.Containers[0].Ports

	for _, port := range ports {
		if port.Name == "grpc" {
			grpcPort = port.ContainerPort
		}
		if port.Name == "https" {
			httpsPort = port.ContainerPort
		}
	}
	return deployment.Spec.Template.Spec.Containers[0].Name, httpsPort, grpcPort, nil
}

func generateService(fn *functionConfig, deploymentPort int32, grpcPort int32) (*yaml.RNode, error) {
	// create a service object over here
	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fn.App,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					Name:       "https",
					TargetPort: intstr.IntOrString(intstr.FromInt(int(deploymentPort))),
				},
			},
			Selector: map[string]string{
				"app": fn.App,
			},
		},
	}
	if grpcPort != 0 {
		service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
			Port:       80,
			Name:       "grpc",
			TargetPort: intstr.IntOrString(intstr.FromInt(int(grpcPort))),
		})
	}
	// append service to the file
	serviceNode, err := fnutils.MakeRNode(service)
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

	certificateNode, err := fnutils.MakeRNode(certificate)
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
