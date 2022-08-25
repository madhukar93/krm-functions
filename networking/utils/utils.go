package utils

import (
	"os"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func CreateService(name string, service v1.Service) error {
	// convert service object to yaml string and then append to file
	svcYaml, err := yaml.Marshal(service)
	if err != nil {
		return err
	}

	err = os.WriteFile(name, svcYaml, 0755)
	if err != nil {
		return err
	}
	return nil
}
