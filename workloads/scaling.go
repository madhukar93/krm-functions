package main

import (
	kedav1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const argoApiVersion = "argoproj.io/v1alpha1"

func Int32(v int32) *int32 {
	return &v
}

func (config functionConfig) makeScaledObject(d appsv1.Deployment) kedav1alpha1.ScaledObject {
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
			MinReplicaCount: Int32(config.Spec.Scaling.MinReplica),
			MaxReplicaCount: Int32(config.Spec.Scaling.MaxReplica),
			Triggers: []kedav1alpha1.ScaleTriggers{
				{
					Type: "gcp-pubsub",
					Metadata: map[string]string{
						"subscriptionName": config.Spec.Scaling.PubsubTopic.Name,
						"subscriptionSize": config.Spec.Scaling.PubsubTopic.Size,
					},
					AuthenticationRef: &kedav1alpha1.ScaledObjectAuthRef{
						Name: "keda-trigger-auth-gcp-credentials",
					},
				},
				{
					Type: "memory",
					Metadata: map[string]string{
						"type":  "Utilization",
						"value": config.Spec.Scaling.Cpu.Target,
					},
				},
				{
					Type: "cpu",
					Metadata: map[string]string{
						"type":  "Utilization",
						"value": config.Spec.Scaling.Cpu.Target,
					},
				},
			},
			ScaleTargetRef: &kedav1alpha1.ScaleTarget{
				APIVersion: argoApiVersion,
				Kind:       "Rollout",
				Name:       d.ObjectMeta.Name,
			},
		},
	}
	return scaledObject
}
