package fnutils

import "sigs.k8s.io/kustomize/kyaml/yaml"

const (
	CONFIG_CONNECTOR_ANNOTATION = "cnrm.cloud.google.com/project-id"
)

func AnnotateConfigConnectorObject(items []*yaml.RNode, project string) ([]*yaml.RNode, error) {
	for i := range items {
		if err := items[i].PipeE(yaml.SetAnnotation(CONFIG_CONNECTOR_ANNOTATION, project)); err != nil {
			return nil, err
		}
	}
	return items, nil
}
