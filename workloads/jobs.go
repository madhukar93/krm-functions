package main

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type jobFunctionConfig struct {
	typeMeta          metav1.TypeMeta
	metav1.ObjectMeta `json:"metadata"`
	Spec              jobSpec `json:"spec"`
}

type jobSpec struct {
	spec
	Schedule string `json:"schedule,omitempty"`
}

func GetJobSpec(jobConf jobFunctionConfig) batchv1.JobSpec {
	jobSpec := batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: jobConf.Spec.App,
				Labels: map[string]string{
					"part-of": jobConf.Spec.PartOf,
					"app":     jobConf.Spec.App,
				},
			},
			Spec: corev1.PodSpec{
				Containers: jobConf.Spec.GetContainers(),
			},
		},
	}
	return jobSpec
}

func GetJobTemplate(jobConf jobFunctionConfig) batchv1.JobTemplateSpec {
	jobTemplateSpec := batchv1.JobTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobConf.Spec.App,
			Labels: map[string]string{
				"part-of": jobConf.Spec.PartOf,
				"app":     jobConf.Spec.App,
			},
		},
		Spec: GetJobSpec(jobConf),
	}
	return jobTemplateSpec
}

func makeCronJob(jobConfig jobFunctionConfig) batchv1.CronJob {
	cj := batchv1.CronJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: jobConfig.Spec.App,
			Labels: map[string]string{
				"part-of": jobConfig.Spec.PartOf,
				"app":     jobConfig.Spec.App,
			},
		},
		Spec: batchv1.CronJobSpec{
			Schedule:    jobConfig.Spec.Schedule,
			JobTemplate: GetJobTemplate(jobConfig),
		},
	}
	return cj
}

func makeJob(jobConfig jobFunctionConfig) batchv1.Job {
	job := batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: jobConfig.Spec.App,
			Labels: map[string]string{
				"part-of": jobConfig.Spec.PartOf,
				"app":     jobConfig.Spec.App,
			},
		},
		Spec: GetJobSpec(jobConfig),
	}
	return job
}
