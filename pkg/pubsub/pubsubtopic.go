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
	Prefix  string   `json:"prefix"`
	Env     string   `json:"env"`
	Topics  []Topic  `json:"topics"`
}

type PubSubTopicConfig struct {
	MessageRetentionDuration string `json:"messageRetentionDuration,omitempty"`
}

type Topic struct {
	TopicName string            `json:"name"`
	Config    PubSubTopicConfig `json:"config,omitempty"`
}

func (p *PubsubTopic) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	out := []*kyaml.RNode{}
	envPrefix := p.Spec.Prefix
	for _, topic := range p.Spec.Topics {
		pubSubTopic := makePubSubTopic(envPrefix+topic.TopicName, topic.Config)
		if pubSubTopic, err := fnutils.MakeRNode(pubSubTopic); err != nil {
			return nil, err
		} else {
			out = append(out, pubSubTopic)
		}
		deadLetterTopic := makePubSubTopic(envPrefix+topic.TopicName+".dlx", topic.Config)
		if deadLetterTopic, err := fnutils.MakeRNode(deadLetterTopic); err != nil {
			return nil, err
		} else {
			out = append(out, deadLetterTopic)
		}
		p := PubSubConfig{}
		deadLetterSubscription := makePubSubSubscription(deadLetterTopic.Name, deadLetterTopic.Name, p)
		if deadLetterSubscription, err := fnutils.MakeRNode(deadLetterSubscription); err != nil {
			return nil, err
		} else {
			out = append(out, deadLetterSubscription)
		}
	}
	out, err := fnutils.AnnotateConfigConnectorObject(out, fnutils.GetProject(p.Spec.Env))
	return out, err
}

func (p PubsubTopic) Schema() (*spec.Schema, error) {
	crdFile, err := ioutil.ReadFile("crd/pubsub/krm_pubsubtopics.yaml")
	schema, err := framework.SchemaFromFunctionDefinition(resid.NewGvk("krm", "v1", "PubsubTopic"), string(crdFile))
	return schema, errors.WrapPrefixf(err, "parsing PubSub schema")
}

func makePubSubTopic(pubSubTopicName string, c PubSubTopicConfig) pubsub.PubSubTopic {
	pubSubTopic := pubsub.PubSubTopic{
		TypeMeta: metav1.TypeMeta{
			Kind:       pubsub.PubSubTopicGVK.Kind,
			APIVersion: pubsub.PubSubTopicGVK.Group + "/" + pubsub.PubSubTopicGVK.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: pubSubTopicName,
		},
		Spec: pubsub.PubSubTopicSpec{
			ResourceID:               &pubSubTopicName,
			MessageRetentionDuration: &c.MessageRetentionDuration,
		},
	}
	return pubSubTopic
}
