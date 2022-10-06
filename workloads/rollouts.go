package main

import (
	rolloutv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type strategy struct {
	AnalysisEnv string `json:"analysis-env"`
}

func IntPtr(x int32) *int32 {
	return &x
}

func (c *functionConfig) addRolloutContainers(r *rolloutv1alpha1.Rollout) error {
	r.Spec.Template.Spec.Containers = append(r.Spec.Template.Spec.Containers, c.Spec.GetContainers()...)
	return nil
}

func (c *functionConfig) addRolloutLabels(r *rolloutv1alpha1.Rollout) error {
	labels := map[string]string{
		"part-of": c.Spec.PartOf,
		"app":     c.Spec.App,
	}
	for k, v := range labels {
		r.Labels[k] = v
		r.Spec.Selector.MatchLabels[k] = v
		r.Spec.Template.Labels[k] = v
	}
	return nil
}

func (s *strategy) setCanarySteps(rollout *rolloutv1alpha1.Rollout) error {
	if s.AnalysisEnv == "prod" {
		rollout.Spec.Strategy.Canary.Steps = []rolloutv1alpha1.CanaryStep{
			{
				SetWeight: IntPtr(30),
			},
			{
				Pause: &rolloutv1alpha1.RolloutPause{
					Duration: &intstr.IntOrString{IntVal: 300},
				},
			},
			{
				SetWeight: IntPtr(60),
			},
			{
				Pause: &rolloutv1alpha1.RolloutPause{
					Duration: &intstr.IntOrString{IntVal: 600},
				},
			},
			{
				SetWeight: IntPtr(100),
			},
		}
	} else if s.AnalysisEnv == "pre-prod" {
		rollout.Spec.Strategy.Canary.Steps = []rolloutv1alpha1.CanaryStep{
			{
				SetWeight: IntPtr(100),
			},
		}
	}

	return nil
}

func (s *strategy) addStrategy(r *rolloutv1alpha1.Rollout) error {

	// TODO : Template name to be fetched dynamically, hardcoded to get started with the canary implementation

	r.Spec.Strategy = rolloutv1alpha1.RolloutStrategy{
		Canary: &rolloutv1alpha1.CanaryStrategy{
			Analysis: &rolloutv1alpha1.RolloutAnalysisBackground{
				RolloutAnalysis: rolloutv1alpha1.RolloutAnalysis{
					Templates: []rolloutv1alpha1.RolloutAnalysisTemplate{
						{
							TemplateName: "analysis-datadog-tokko-api-graphql-error-rate",
						},
					},
					Args: []rolloutv1alpha1.AnalysisRunArgument{
						{
							Name: "service-name",
							ValueFrom: &rolloutv1alpha1.ArgumentValueFrom{
								FieldRef: &rolloutv1alpha1.FieldRef{
									FieldPath: "metadata.name",
								},
							},
						},
						{
							Name: "env",
							ValueFrom: &rolloutv1alpha1.ArgumentValueFrom{
								FieldRef: &rolloutv1alpha1.FieldRef{
									FieldPath: "metadata.annotations['app.tokko.io/env']",
								},
							},
						},
						{
							Name: "version",
							ValueFrom: &rolloutv1alpha1.ArgumentValueFrom{
								FieldRef: &rolloutv1alpha1.FieldRef{
									FieldPath: "metadata.annotations['app.tokko.io/version']",
								},
							},
						},
					},
				},
				StartingStep: IntPtr(2),
			},
		},
	}
	return nil
}

func NewRollout() *rolloutv1alpha1.Rollout {
	rollout := rolloutv1alpha1.Rollout{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Rollout",
			APIVersion: "argoproj.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "",
			Namespace: "",
			Labels:    map[string]string{},
		},
		Spec: rolloutv1alpha1.RolloutSpec{
			Replicas: nil,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{},
				},
			},
			Strategy: rolloutv1alpha1.RolloutStrategy{},
		},
	}
	return &rollout
}

func makeRollout(conf functionConfig) rolloutv1alpha1.Rollout {
	rollout := NewRollout()
	rollout.ObjectMeta.Name = conf.Spec.App
	conf.addRolloutContainers(rollout)
	conf.addRolloutLabels(rollout)
	conf.Spec.Strategy.addStrategy(rollout)
	conf.Spec.Strategy.setCanarySteps(rollout)
	return *rollout
}
