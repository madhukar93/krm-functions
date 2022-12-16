package fnutils

import (
	"encoding/json"
	"fmt"
	"os"

	esapi "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
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

// MakeRnodes converts k8s API objects into RNodes for KRM functions
func MakeRNodes(objects ...metav1.Object) ([]*kyaml.RNode, error) {
	var res framework.Results
	var out []*kyaml.RNode

	for _, o := range objects {
		result := &framework.Result{}
		result.ResourceRef = &kyaml.ResourceIdentifier{
			NameMeta: kyaml.NameMeta{
				Name:      o.GetName(),
				Namespace: o.GetNamespace(),
			},
		}
		u := &unstructured.Unstructured{}
		if uc, err := runtime.DefaultUnstructuredConverter.ToUnstructured(o); err != nil {
			result.Severity = framework.Error
			result.Message = fmt.Sprintf("failed to convert to unstructured: %v", err)
		} else {
			// only able to get TypeMeta from unstructured
			u.SetUnstructuredContent(uc)
			result.ResourceRef.TypeMeta = kyaml.TypeMeta{
				APIVersion: u.GetAPIVersion(),
				Kind:       u.GetKind(),
			}

			if rNode, err := kyaml.FromMap(uc); err != nil {
				result.Severity = framework.Error
				result.Message = fmt.Sprintf("failed to convert to RNode: %v", err)
			} else {
				out = append(out, rNode)
				result.Message = "successully converted to RNode"
				result.Severity = framework.Info
			}
		}

	}
	return out, res
}

// Function that parses the RNode into a ExternalSecret
func ParseRNodeExternalSecret(item *kyaml.RNode) (*esapi.ExternalSecret, error) {
	var secret esapi.ExternalSecret
	// Convert RNode to JSON
	jsonBytes, err := item.MarshalJSON()
	if err != nil {
		return nil, err
	}
	// Convert JSON to ExternalSecret
	if err := json.Unmarshal(jsonBytes, &secret); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	return &secret, nil
}
