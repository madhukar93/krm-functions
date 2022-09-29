package fnutils

import (
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
)

// MakeRNode creates a RNode from yaml Marshallable object
func MakeRNode(in any) (*kyaml.RNode, error) {
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
