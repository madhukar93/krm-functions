package main

import (
	"errors"
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/bukukasio/krm-functions/pkg/fnutils"
	cm "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	traefik "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	ingressRouteKind     = "IngressRoute"
	apiVersionNetworking = "traefik.containo.us/v1alpha1"
)

type RouteConfig struct {
	// TODO: can probably just use only json tags
	Match string `yaml:"match ,omitempty" ,json:"match, omitempty"`
	Vpn   bool   `yaml:"vpn ,omitempty" ,json:"vpn ,omitempty"`
}

type functionConfig struct {
	App    string        `yaml:"app" ,json:"app"`
	Hosts  []string      `yaml:"hosts" ,json:"hosts"`
	Grpc   bool          `yaml:"grpc ,omitempty" ,json:"grpc ,omitempty"`
	Routes []RouteConfig `yaml:"routes" ,json:"routes"`
}

func (fnConfig *functionConfig) Filter(items []*kyaml.RNode) ([]*kyaml.RNode, error) {
	out := []*kyaml.RNode{}
	var err error
	var wlNetwConf *workloadNetworkConfig
	wlNetwConf, err = workloadNetworking(deploymentNode)

	// check for deployment with app label
	foundDeployment := true
	for _, item := range items {
		meta, err := item.GetMeta()
		meta.
		if err != nil {
			return items, err
		}

		if meta.Kind != "Service" && meta.Kind != "IngressRoute" && meta.Kind != "Certificate" {
			out = append(out, resource)
		}
	}

	if !foundDeployment {
		return items, fmt.Errorf("could not find deployment with name: %s", fn.App)
	}

	for _, resource := range items {
		meta, err := resource.GetMeta()
		if err != nil {
			return items, err
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

type workloadNetworkConfig struct {
	containerName string
	grpcPort int32
	httpPort int32
}

func workloadNetworking(node *kyaml.RNode) (*workloadNetworkConfig, error) {
	deployment := appsv1.Deployment{}
	dYaml, err := yaml.Marshal(node)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(dYaml, &deployment); err != nil {
		return nil, err
	}

	var httpPort int32
	var grpcPort int32

	ports := deployment.Spec.Template.Spec.Containers[0].Ports

	for _, port := range ports {
		if port.Name == "grpc" {
			grpcPort = port.ContainerPort
		}
		if port.Name == "https" {
			httpPort = port.ContainerPort
		}
	}
	return &workloadNetworkConfig{
		deployment.Spec.Template.Spec.Containers[0].Name,
		httpPort,
		grpcPort,
		},  nil
}

func generateService(fn *functionConfig, deploymentPort int32, grpcPort int32) (*kyaml.RNode, error) {
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

func generateCertificate(fn *functionConfig) (*kyaml.RNode, error) {

	certificate := cm.Certificate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Certificate",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fn.App,
		},
		Spec: cm.CertificateSpec{
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
