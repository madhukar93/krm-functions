package main

import (
	"fmt"

	"github.com/bukukasio/krm-functions/pkg/fnutils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type functionConfig struct {
	metav1.TypeMeta
	metav1.ObjectMeta `json:"metadata"`
	Spec              spec `json:"spec"`
}

type spec struct {
	PartOf     string       `json:"part-of"`
	App        string       `json:"app"`
	Env        string       `json:"env,omitempty"`
	Containers []container  `json:"containers,omitempty"`
	Scaling    *scalingSpec `json:"scaling,omitempty"`
	Strategy   *strategy    `json:"strategy,omitempty"`
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

func (fnConfig *functionConfig) Filter(nodes []*kyaml.RNode) ([]*kyaml.RNode, error) {
	out := []*kyaml.RNode{}
	if fnConfig.Kind == "LummoDeployment" {
		deployment := makeDeployment(*fnConfig)
		service := makeService(deployment)
		if d, err := fnutils.MakeRNode(deployment); err != nil {
			return nil, err
		} else {
			out = append(out, d)
		}
		if s, err := fnutils.MakeRNode(service); err != nil {
			return nil, err
		} else {
			out = append(out, s)
		}
		if fnConfig.Spec.Scaling != nil {
			scaling := fnConfig.Spec.Scaling.makeScaledObject(deployment)
			if s, err := fnutils.MakeRNode(scaling); err != nil {
				return nil, err
			} else {
				out = append(out, s)
			}
		}
	}

	if fnConfig.Kind == "LummoRollout" {
		rollout := makeRollout(*fnConfig)
		if d, err := fnutils.MakeRNode(rollout); err != nil {
			return nil, err
		} else {
			out = append(out, d)
		}
	}
	return out, nil
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

func NewDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "",
			Namespace: "",
			Labels:    map[string]string{},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: nil,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{},
				},
			},
		},
	}
}

// TODO: Return a Result type? Does the framework have a result type?
func makeDeployment(conf functionConfig) appsv1.Deployment {
	d := NewDeployment()
	d.ObjectMeta.Name = conf.Spec.App
	conf.addDeploymentLabels(d)
	conf.addContainers(d)
	return *d
}

func makeService(d appsv1.Deployment) corev1.Service {
	s := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: d.ObjectMeta.Labels["app"],
			Labels: map[string]string{
				"part-of": d.ObjectMeta.Labels["part-of"],
				"app":     d.ObjectMeta.Labels["app"],
			},
		},
	}
	s.Spec.Selector = d.Spec.Selector.MatchLabels
	// actually should happen for all containers? One service per deployment or container? How do services scale
	// but ingress is probably only needed for app cotainer
	ac, err := getAppContainer(d)

	if err != nil {
		// TODO validation error
		panic(err)
	}

	for _, p := range ac.Ports {
		s.Spec.Ports = append(s.Spec.Ports, corev1.ServicePort{
			Name:       p.Name,
			Port:       p.ContainerPort,
			TargetPort: intstr.FromInt(int(p.ContainerPort)),
		})
	}
	return s
}
