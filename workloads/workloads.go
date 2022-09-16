package main

import (
	"encoding/json"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	PartOf            string      `json:"part-of"`
	App               string      `json:"app"`
	Containers        []container `json:"containers,omitempty"`
}

func (s spec) GetContainers() []corev1.Container {
	// TODO: what was that receiver param thingy? is using that considered good practise?
	cs := []corev1.Container{}
	for _, c := range s.Containers {
		cs = append(cs, c.Container)
	}
	return cs
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
	Configs []string `json:"configs"`
	Secrets []string `json:"secrets"`
	Volumes []string `json:"volumes"`
}

func (c *container) GetContainer() corev1.Container {
	// TODO process extra fields
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
				if replacement, err := makeRNode(deployment); err != nil {
					return nil, err
				} else {
					out = append(out, replacement)
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

// deployment builds a appsv1.Deployment from the functionConfig
// TODO: validate
func makeDeployment(conf functionConfig) appsv1.Deployment {
	return appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   conf.Spec.App,
			//Labels: conf.Spec.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     conf.Spec.App,
					"part-of": conf.Spec.PartOf, // TODO: these are ORs, not ANDs to confirm
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.PodSpec{
					Containers: conf.Spec.GetContainers(),
				},
			},
		},
	}
}
