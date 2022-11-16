package testing

import (
	"github.com/wI2L/jsondiff"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Yaml wrapper for jsondiff
func YamlDiff(first []byte, second []byte) (*jsondiff.Patch, error) {
	var err error

	a, err := yaml.ToJSON(first)
	if err != nil {
		return nil, err
	}

	b, err := yaml.ToJSON(second)
	if err != nil {
		return nil, err
	}

	diff, err := jsondiff.CompareJSON(a, b)
	if err != nil {
		return nil, err
	}
	return &diff, nil
}
