// +kubebuilder:object:generate=true
// +groupName=krm
// +versionName=v1
package iam

import (
	"io/ioutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/parser"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// +kubebuilder:object:root=true
type LummoIAM struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              Spec `json:"spec"`
}

type Spec struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Project   string `json:"project"`
}

func (a LummoIAM) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	filter := framework.TemplateProcessor{
		ResourceTemplates: []framework.ResourceTemplate{{
			TemplateData: &a,
			Templates:    parser.TemplateFiles("templates/iam/iam.template.yaml"),
		}},
	}
	return filter.Filter(items)
}

func (a LummoIAM) Schema() (*spec.Schema, error) {
	crdFile, err := ioutil.ReadFile("crd/iam/krm_lummoiams.yaml")
	schema, err := framework.SchemaFromFunctionDefinition(resid.NewGvk("krm", "v1", "LummoIAM"), string(crdFile))
	return schema, errors.WrapPrefixf(err, "parsing IAM schema")
}
