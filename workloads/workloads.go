package main

import (
	"encoding/json"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
)

type functionConfig struct {
	typeMeta          metav1.TypeMeta
	metav1.ObjectMeta `json:"metadata"`
	Spec              spec `json:"spec"`
}

type spec struct {
	PartOf     string      `json:"part-of"`
	App        string      `json:"app"`
	Containers []container `json:"containers,omitempty"`
}

func (s spec) GetContainers() []corev1.Container {
	// TODO: what was that receiver param thingy? is using that considered good practise?
	cs := []corev1.Container{}
	for _, c := range s.Containers {
		cs = append(cs, c.GetContainer())
	}
	return cs
}

type grpc struct {
	Port int32 `json:"port"`
}

func (p *grpc) setGrpcPort(c *corev1.Container) error {
	port := corev1.ContainerPort{
		Name:          "grpc",
		ContainerPort: p.Port,
		Protocol:      corev1.ProtocolTCP,
	}
	c.Ports = append(c.Ports, port)
	return nil
}

type http struct {
	Port int32 `json:"port"`
}

func (p *http) setHttpPort(c *corev1.Container) error {
	port := corev1.ContainerPort{
		Name:          "http",
		ContainerPort: p.Port,
		Protocol:      corev1.ProtocolTCP,
	}
	c.Ports = append(c.Ports, port)
	return nil
}

type config string

func (c config) envFromConfigMap() corev1.EnvFromSource {
	return corev1.EnvFromSource{
		ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: string(c),
			},
		},
	}
}

type secret string

func (s secret) envFromSecret() corev1.EnvFromSource {
	return corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: string(s),
			},
		},
	}
}

type container struct {
	corev1.Container `json:",inline"`
	// this is docker compose-ish
	// if these fields are populated, they augment the container
	// if they are not populated, the container is used as is
	// if they are populated, the container is used as a base
	// and the fields are applied on top
	// if they are populated, the container is used as a base
	// and the fields are applied on top
	Configs []config `json:"configs"`
	Secrets []secret `json:"secrets"`
	Grpc    grpc     `json:"grpc,omitempty"`
	Http    http     `json:"http,omitempty"`

	// Complicated because of PVC stuff and not worth doing
	//Volumes []string `json:"volumes"`
}

func (c *container) GetContainer() corev1.Container {
	// TODO process extra fields
	for _, config := range c.Configs {
		c.EnvFrom = append(c.EnvFrom, config.envFromConfigMap())
	}
	for _, secret := range c.Secrets {
		c.EnvFrom = append(c.EnvFrom, secret.envFromSecret())
	}

	if c.Grpc.Port != 0 {
		c.Grpc.setGrpcPort(&c.Container)
	}
	if c.Http.Port != 0 {
		c.Http.setHttpPort(&c.Container)
	}
	return c.Container
}

type WorkloadsFilter struct {
	rw *kio.ByteReadWriter
}

func (w WorkloadsFilter) Filter(nodes []*kyaml.RNode) ([]*kyaml.RNode, error) {
	out := []*kyaml.RNode{}
	// TODO: use switch
	for _, node := range nodes {
		if node.GetKind() == "Deployment" {
			continue

		} else if node.GetKind() == "LummoDeployment" {
			if fnConfig, err := parseFnConfig(node); err != nil {
				return nil, err
			} else {
				deployment := makeDeployment(*fnConfig)
				service := makeService(deployment)
				if d, err := makeRNode(deployment); err != nil {
					return nil, err
				} else {
					out = append(out, d)
				}
				if s, err := makeRNode(service); err != nil {
					return nil, err
				} else {
					out = append(out, s)
				}
			}
			continue
		}

		out = append(out, node)
	}

	return out, nil
}

// makeRNode creates a RNode from yaml Marshallable object
func makeRNode(in any) (*kyaml.RNode, error) {
	if yml, err := yaml.Marshal(in); err != nil {
		return nil, err
	} else {
		if rnode, err := kyaml.Parse(string(yml)); err != nil {
			return nil, err
		} else {
			return rnode, nil
		}
	}
}

// parseFnConfig parses the functionConfig into the functionConfig struct
func parseFnConfig(node *kyaml.RNode) (*functionConfig, error) {
	var config functionConfig
	jsonBytes, err := node.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	return &config, nil
}

// validations
// If there are ports exposed, there must be a service and probes on that port for the 'app' container

// deployment builds a appsv1.Deployment from the functionConfig
// TODO: validate

func (c *functionConfig) addDeploymentLabels(d *appsv1.Deployment) error {
	labels := map[string]string{
		"part-of": c.Spec.PartOf,
		"app":     c.Spec.App,
	}
	for k, v := range labels {
		d.Labels[k] = v
		d.Spec.Selector.MatchLabels[k] = v
		d.Spec.Template.Labels[k] = v
	}
	return nil
}

func (c *functionConfig) addContainers(d *appsv1.Deployment) error {
	d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers, c.Spec.GetContainers()...)
	return nil
}

func getAppContainer(d appsv1.Deployment) (*corev1.Container, error) {
	for _, c := range d.Spec.Template.Spec.Containers {
		if c.Name == d.Labels["app"] {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("no app container found")
}

// TODO: Return a Result type? Does the framework have a result type?
func makeDeployment(conf functionConfig) appsv1.Deployment {
	d := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: conf.Spec.App,
			//Labels: conf.Spec.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       corev1.PodSpec{},
			},
		},
	}

	d.Labels = make(map[string]string)
	d.Spec.Selector.MatchLabels = make(map[string]string)
	d.Spec.Template.Labels = make(map[string]string)
	conf.addDeploymentLabels(&d)
	d.Spec.Template.Spec.Containers = append(conf.Spec.GetContainers(), d.Spec.Template.Spec.Containers...)
	return d
}

func makeService(d appsv1.Deployment) corev1.Service {
	s := corev1.Service{}
	s.Spec.Selector = d.Spec.Selector.MatchLabels
	ac, _ := getAppContainer(d)
	for _, p := range ac.Ports {
		s.Spec.Ports = append(s.Spec.Ports, corev1.ServicePort{
			Name:       p.Name,
			Port:       p.ContainerPort,
			TargetPort: intstr.FromInt(int(p.ContainerPort)),
		})
	}
	return s
}
