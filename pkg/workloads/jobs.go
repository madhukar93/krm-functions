package workloads

import (
	"github.com/bukukasio/krm-functions/pkg/common/fnutils"
	utils "github.com/bukukasio/krm-functions/pkg/common/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type JobFunctionConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              jobSpec `json:"spec"`
}

type jobSpec struct {
	podSpec            `json:",inline"`
	RestartPolicy      string `json:"restartPolicy,omitempty"`
	Schedule           string `json:"schedule,omitempty"`
	GenerateNameSuffix bool   `json:"generateNameSuffix,omitempty"`
}

func GetJobSpec(jobConf JobFunctionConfig) batchv1.JobSpec {
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
				Containers:    jobConf.Spec.GetContainers(),
				RestartPolicy: corev1.RestartPolicy(jobConf.Spec.RestartPolicy),
			},
		},
	}
	return jobSpec
}

func GetJobTemplate(jobConf JobFunctionConfig) batchv1.JobTemplateSpec {
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

func makeCronJob(jobConfig JobFunctionConfig) batchv1.CronJob {
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

func makeJob(jobConfig JobFunctionConfig) batchv1.Job {
	var name string
	if jobConfig.Spec.GenerateNameSuffix {
		name = jobConfig.Spec.App + "-" + utils.RandomString(8)
	} else {
		name = jobConfig.Spec.App
	}
	job := batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"part-of": jobConfig.Spec.PartOf,
				"app":     jobConfig.Spec.App,
			},
		},
		Spec: GetJobSpec(jobConfig),
	}
	return job
}

func (fnConfig *JobFunctionConfig) Filter(nodes []*kyaml.RNode) ([]*kyaml.RNode, error) {
	out := []*kyaml.RNode{}
	if fnConfig.Kind == "LummoJob" {
		job := makeJob(*fnConfig)
		if d, err := fnutils.MakeRNode(job); err != nil {
			return nil, err
		} else {
			out = append(out, d)
		}
	}

	if fnConfig.Kind == "LummoCron" {
		cronjob := makeCronJob(*fnConfig)
		if d, err := fnutils.MakeRNode(cronjob); err != nil {
			return nil, err
		} else {
			out = append(out, d)
		}
	}
	return out, nil
}

func (a JobFunctionConfig) Schema() (*spec.Schema, error) {
	schema, err := framework.SchemaFromFunctionDefinition(resid.NewGvk("krm", "workloads", "JobFunctionConfig"), fnutils.LoadConfig("crd/workloads/krm_jobfunctionconfigs.yaml"))
	return schema, errors.WrapPrefixf(err, "\n parsing jobs schema")
}
