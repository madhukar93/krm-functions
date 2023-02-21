// +kubebuilder:object:generate=true
// +groupName=krm
// +versionName=v1
package pubsub

import (
	"io/ioutil"

	pubsub "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/clients/generated/apis/pubsub/v1beta1"
	"github.com/bukukasio/krm-functions/pkg/common/fnutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// +kubebuilder:object:root=true
type PubsubTopic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PubsubTopicSpec `json:"spec"`
}

type PubsubTopicSpec struct {
	Prefix string   `json:"prefix"`
	Topics []string `json:"topics"`
}

func (p *PubsubTopic) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	out := []*kyaml.RNode{}
	envPrefix, ok := p.ObjectMeta.Labels["env"]
	if !ok {
		return nil, errors.Errorf("env label not found! Can't mutate pubsub topic")
	}
	for _, topic := range p.Spec.Topics {
		pubSubTopic := makePubSubTopic(envPrefix + "-" + topic)
		if pubSubTopic, err := fnutils.MakeRNode(pubSubTopic); err != nil {
			return nil, err
		} else {
			out = append(out, pubSubTopic)
		}
		deadLetterTopic := makePubSubTopic(envPrefix + "-" + topic + ".dlx")
		if deadLetterTopic, err := fnutils.MakeRNode(deadLetterTopic); err != nil {
			return nil, err
		} else {
			out = append(out, deadLetterTopic)
		}
		p := PubSubConfig{}
		deadLetterSubscription := makePubSubSubscription(envPrefix+"-"+deadLetterTopic.Name, envPrefix+"-"+deadLetterTopic.Name, p)
		if deadLetterSubscription, err := fnutils.MakeRNode(deadLetterSubscription); err != nil {
			return nil, err
		} else {
			out = append(out, deadLetterSubscription)
		}
	}
	return out, nil
}

func (p PubsubTopic) Schema() (*spec.Schema, error) {
	crdFile, err := ioutil.ReadFile("crd/pubsub/krm_pubsubtopics.yaml")
	schema, err := framework.SchemaFromFunctionDefinition(resid.NewGvk("krm", "v1", "PubsubTopic"), string(crdFile))
	return schema, errors.WrapPrefixf(err, "parsing PubSub schema")
}

func makePubSubTopic(pubSubTopicName string) pubsub.PubSubTopic {
	pubSubTopic := pubsub.PubSubTopic{
		TypeMeta: metav1.TypeMeta{
			Kind:       pubsub.PubSubTopicGVK.Kind,
			APIVersion: pubsub.PubSubTopicGVK.Group + "/" + pubsub.PubSubTopicGVK.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: pubSubTopicName,
		},
		Spec: pubsub.PubSubTopicSpec{
			ResourceID: &pubSubTopicName,
		},
	}
	return pubSubTopic
}
