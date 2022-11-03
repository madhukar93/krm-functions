package workloads

import (
	kedav1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const argoApiVersion = "argoproj.io/v1alpha1"
const kedaTriggerAuth = "keda-trigger-auth-gcp-credentials"

type cpu struct {
	Target string `json:"target,omitempty"`
}

type memory struct {
	Target string `json:"target,omitempty"`
}

type pubsubTopic struct {
	Name string `json:"name,omitempty"`
	Size string `json:"size,omitempty"`
}

type scalingSpec struct {
	MinReplica  int32        `json:"minreplica"`
	MaxReplica  int32        `json:"maxreplica"`
	Cpu         *cpu         `json:"cpu,omitempty"`
	Memory      *memory      `json:"memory,omitempty"`
	PubsubTopic *pubsubTopic `json:"pubsubTopic,omitempty"`
}

func (spec scalingSpec) addCpuTrigger(so *kedav1alpha1.ScaledObject) {
	cpuTrigger := []kedav1alpha1.ScaleTriggers{
		{
			Type: "cpu",
			Metadata: map[string]string{
				"type":  "Utilization",
				"value": spec.Cpu.Target,
			},
		},
	}
	so.Spec.Triggers = append(so.Spec.Triggers, cpuTrigger...)
}

func (spec scalingSpec) addMemoryTrigger(so *kedav1alpha1.ScaledObject) {
	memoryTrigger := []kedav1alpha1.ScaleTriggers{
		{
			Type: "memory",
			Metadata: map[string]string{
				"type":  "Utilization",
				"value": spec.Memory.Target,
			},
		},
	}
	so.Spec.Triggers = append(so.Spec.Triggers, memoryTrigger...)
}

func (spec scalingSpec) addPubSubTrigger(so *kedav1alpha1.ScaledObject) {
	pubsubTrigger := []kedav1alpha1.ScaleTriggers{
		{
			Type: "gcp-pubsub",
			Metadata: map[string]string{
				"subscriptionName": spec.PubsubTopic.Name,
				"subscriptionSize": spec.PubsubTopic.Size,
			},
			AuthenticationRef: &kedav1alpha1.ScaledObjectAuthRef{
				Name: kedaTriggerAuth,
			},
		},
	}
	so.Spec.Triggers = append(so.Spec.Triggers, pubsubTrigger...)
}

func (spec scalingSpec) makeScaledObject(d appsv1.Deployment) kedav1alpha1.ScaledObject {
	scaledObject := kedav1alpha1.ScaledObject{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ScaledObject",
			APIVersion: "keda.sh/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: d.ObjectMeta.Name,
			Labels: map[string]string{
				"part-of": d.ObjectMeta.Labels["part-of"],
				"app":     d.ObjectMeta.Labels["app"],
			},
		},
		Spec: kedav1alpha1.ScaledObjectSpec{
			MinReplicaCount: &spec.MinReplica,
			MaxReplicaCount: &spec.MaxReplica,
			Triggers:        []kedav1alpha1.ScaleTriggers{},
			ScaleTargetRef: &kedav1alpha1.ScaleTarget{
				APIVersion: argoApiVersion,
				Kind:       "Rollout",
				Name:       d.ObjectMeta.Name,
			},
		},
	}
	if spec.Cpu != nil {
		spec.addCpuTrigger(&scaledObject)
	}
	if spec.Memory != nil {
		spec.addMemoryTrigger(&scaledObject)
	}
	if spec.PubsubTopic != nil {
		spec.addPubSubTrigger(&scaledObject)
	}
	return scaledObject
}
