package workloads

import (
	"fmt"

	kedav1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const argoApiVersion = "argoproj.io/v1alpha1"
const kedaTriggerAuth = "keda-trigger-auth-gcp-credentials"
const kedaApiVersion = "keda.sh/v1alpha1"

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

func (spec scalingSpec) makeScaledObject(workloadData metav1.Object) kedav1alpha1.ScaledObject {

	u := &unstructured.Unstructured{}
	if uc, err := runtime.DefaultUnstructuredConverter.ToUnstructured(workloadData); err != nil {
		fmt.Printf("failed to convert to unstructured: %v", err)
	} else {
		u.SetUnstructuredContent(uc)
	}

	scaledObject := kedav1alpha1.ScaledObject{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ScaledObject",
			APIVersion: kedaApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: u.GetName(),
			Labels: map[string]string{
				"part-of": u.GetLabels()["part-of"],
				"app":     u.GetLabels()["app"],
			},
		},
		Spec: kedav1alpha1.ScaledObjectSpec{
			MinReplicaCount: &spec.MinReplica,
			MaxReplicaCount: &spec.MaxReplica,
			Triggers:        []kedav1alpha1.ScaleTriggers{},
			ScaleTargetRef: &kedav1alpha1.ScaleTarget{
				APIVersion: u.GetAPIVersion(),
				Kind:       u.GetKind(),
				Name:       u.GetName(),
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
