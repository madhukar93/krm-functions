// +kubebuilder:object:generate=true
// +groupName=krm
// +versionName=v1
package pubsub

import (
	"io/ioutil"
	"strings"

	resource_ref "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/clients/generated/apis/k8s/v1alpha1"
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
type PubsubSubscription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PubsubSubscriptionSpec `json:"spec"`
}

type PubsubSubscriptionSpec struct {
	Prefix        string         `json:"prefix"`
	Subscriptions []Subscription `json:"subscriptions"`
}

type PubSubConfig struct {
	AckDeadlineSeconds       int    `json:"ackDeadlineSeconds,omitempty"`
	MaxDeliveryAttempts      int    `json:"maxDeliveryAttempts,omitempty"`
	TTL                      string `json:"ttl,omitempty"`
	MessageRetentionDuration string `json:"messageRetentionDuration,omitempty"`
	MaximumBackoff           string `json:"maximumBackoff,omitempty"`
	MinimumBackoff           string `json:"minimumBackoff,omitempty"`
}

type Subscription struct {
	Subscription string       `json:"subscription"`
	TopicRef     string       `json:"topic"`
	Config       PubSubConfig `json:"config,omitempty"`
}

func (pubSubConfig *PubSubConfig) fill_defaults() {
	if pubSubConfig.AckDeadlineSeconds == 0 {
		pubSubConfig.AckDeadlineSeconds = 10
	}
	if pubSubConfig.MaxDeliveryAttempts == 0 {
		pubSubConfig.MaxDeliveryAttempts = 5
	}
	if pubSubConfig.TTL == "" {
		pubSubConfig.TTL = "2678400s"
	}
	if pubSubConfig.MessageRetentionDuration == "" {
		pubSubConfig.MessageRetentionDuration = "604800s"
	}
	if pubSubConfig.MinimumBackoff == "" {
		pubSubConfig.MinimumBackoff = "300s"
	}
	if pubSubConfig.MaximumBackoff == "" {
		pubSubConfig.MaximumBackoff = "600s"
	}
}

func (p *PubsubSubscription) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	out := []*kyaml.RNode{}
	for _, sub := range p.Spec.Subscriptions {
		pubSubTopic := makePubSubSubscription(sub.Subscription, sub.TopicRef, sub.Config)
		if pubSubTopic, err := fnutils.MakeRNode(pubSubTopic); err != nil {
			return nil, err
		} else {
			out = append(out, pubSubTopic)
		}
	}
	return out, nil
}

func (p PubsubSubscription) Schema() (*spec.Schema, error) {
	crdFile, err := ioutil.ReadFile("crd/pubsub/krm_pubsubsubscriptions.yaml")
	schema, err := framework.SchemaFromFunctionDefinition(resid.NewGvk("krm", "v1", "PubsubSubscription"), string(crdFile))
	return schema, errors.WrapPrefixf(err, "parsing PubSub schema")
}

func makePubSubSubscription(pubSubSubscriptionName string, pubSubTopic string, p PubSubConfig) pubsub.PubSubSubscription {
	p.fill_defaults()
	pubSubScription := pubsub.PubSubSubscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       pubsub.PubSubSubscriptionGVK.Kind,
			APIVersion: pubsub.PubSubSubscriptionGVK.Group + "/" + pubsub.PubSubSubscriptionGVK.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: pubSubSubscriptionName,
		},
		Spec: pubsub.PubSubSubscriptionSpec{
			TopicRef: resource_ref.ResourceRef{
				Name: pubSubTopic,
			},
			MessageRetentionDuration: &p.MessageRetentionDuration,
			AckDeadlineSeconds:       &p.AckDeadlineSeconds,
			ExpirationPolicy: &pubsub.SubscriptionExpirationPolicy{
				Ttl: p.TTL,
			},
		},
	}
	if !strings.HasSuffix(pubSubSubscriptionName, "dlx") {
		pubSubScription.Spec.DeadLetterPolicy = &pubsub.SubscriptionDeadLetterPolicy{
			DeadLetterTopicRef: &resource_ref.ResourceRef{
				Name: pubSubTopic + ".dlx",
			},
			MaxDeliveryAttempts: &p.MaxDeliveryAttempts,
		}
		pubSubScription.Spec.RetryPolicy = &pubsub.SubscriptionRetryPolicy{
			MaximumBackoff: &p.MaximumBackoff,
			MinimumBackoff: &p.MinimumBackoff,
		}
	}
	return pubSubScription
}
