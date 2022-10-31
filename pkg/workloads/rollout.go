package workloads

import (
	rolloutv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type strategy struct {
	AnalysisMetrics metrics `json:"metrics"`
}

type metrics struct {
	Datadog datadog `json:"datadog"`
}

type datadog struct {
	Operation  string  `json:"operation"`
	ErrorRPM   *string `json:"errorRPM,omitempty"`
	P95latency *string `json:"p95latency,omitempty"`
}

func PointerTo[T any](v T) *T {
	return &v
}

func (c *FunctionConfig) addRolloutContainers(r *rolloutv1alpha1.Rollout) {
	r.Spec.Template.Spec.Containers = append(r.Spec.Template.Spec.Containers, c.Spec.GetContainers()...)
}

func (c *FunctionConfig) addRolloutLabels(r *rolloutv1alpha1.Rollout) {
	labels := map[string]string{
		"part-of": c.Spec.PartOf,
		"app":     c.Spec.App,
	}
	for k, v := range labels {
		r.Labels[k] = v
		r.Spec.Selector.MatchLabels[k] = v
		r.Spec.Template.Labels[k] = v
	}
}

func (s *strategy) setCanarySteps(rollout *rolloutv1alpha1.Rollout, env string) {
	if env == "prod" {
		rollout.Spec.Strategy.Canary.Steps = []rolloutv1alpha1.CanaryStep{
			{
				SetWeight: PointerTo[int32](30),
			},
			{
				Pause: &rolloutv1alpha1.RolloutPause{
					Duration: &intstr.IntOrString{IntVal: 300},
				},
			},
			{
				SetWeight: PointerTo[int32](60),
			},
			{
				Pause: &rolloutv1alpha1.RolloutPause{
					Duration: &intstr.IntOrString{IntVal: 600},
				},
			},
			{
				SetWeight: PointerTo[int32](100),
			},
		}
	} else {
		rollout.Spec.Strategy.Canary.Steps = []rolloutv1alpha1.CanaryStep{
			{
				SetWeight: PointerTo[int32](100),
			},
		}
	}
}

func getAnalysisTemplate(template string) rolloutv1alpha1.RolloutAnalysisTemplate {
	rolloutTemplate := rolloutv1alpha1.RolloutAnalysisTemplate{
		TemplateName: template,
	}
	return rolloutTemplate
}

func (s *strategy) addAnalysisTemplates(r *rolloutv1alpha1.Rollout) {
	var templates_list []string
	if s.AnalysisMetrics.Datadog.ErrorRPM != nil {
		templates_list = append(templates_list, "analysis-datadog-request-errors")
	}
	if s.AnalysisMetrics.Datadog.P95latency != nil {
		templates_list = append(templates_list, "analysis-datadog-request-p95-latency")
	}

	for _, template := range templates_list {
		r.Spec.Strategy.Canary.Analysis.RolloutAnalysis.Templates = append(r.Spec.Strategy.Canary.Analysis.RolloutAnalysis.Templates, getAnalysisTemplate(template))
	}
}

func getTemplateArg(argName string, argValue string) rolloutv1alpha1.AnalysisRunArgument {
	arg := rolloutv1alpha1.AnalysisRunArgument{
		Name:  argName,
		Value: argValue,
	}
	return arg
}

func (s *strategy) addTemplateArgs(r *rolloutv1alpha1.Rollout) {

	if s.AnalysisMetrics.Datadog.P95latency != nil {
		r.Spec.Strategy.Canary.Analysis.RolloutAnalysis.Args = append(r.Spec.Strategy.Canary.Analysis.RolloutAnalysis.Args, getTemplateArg("p95latency", *s.AnalysisMetrics.Datadog.P95latency))
	}

	if s.AnalysisMetrics.Datadog.ErrorRPM != nil {
		r.Spec.Strategy.Canary.Analysis.RolloutAnalysis.Args = append(r.Spec.Strategy.Canary.Analysis.RolloutAnalysis.Args, getTemplateArg("errorRPM", *s.AnalysisMetrics.Datadog.ErrorRPM))

	}
}

func (s *strategy) addStrategy(r *rolloutv1alpha1.Rollout) {

	r.Spec.Strategy = rolloutv1alpha1.RolloutStrategy{
		Canary: &rolloutv1alpha1.CanaryStrategy{
			Analysis: &rolloutv1alpha1.RolloutAnalysisBackground{
				RolloutAnalysis: rolloutv1alpha1.RolloutAnalysis{
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
						{
							Name:  "operation",
							Value: s.AnalysisMetrics.Datadog.Operation,
						},
					},
				},
				StartingStep: PointerTo[int32](2),
			},
		},
	}
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

func makeRollout(conf FunctionConfig) rolloutv1alpha1.Rollout {
	rollout := NewRollout()
	rollout.ObjectMeta.Name = conf.Spec.App
	conf.addRolloutContainers(rollout)
	conf.addRolloutLabels(rollout)
	conf.Spec.Strategy.addStrategy(rollout)
	conf.Spec.Strategy.setCanarySteps(rollout, conf.Spec.Env)
	conf.Spec.Strategy.addAnalysisTemplates(rollout)
	conf.Spec.Strategy.addTemplateArgs(rollout)
	return *rollout
}
