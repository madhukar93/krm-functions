// +kubebuilder:object:generate=true
// +groupName=krm
package pgbouncer

import (
	"fmt"
	"io/ioutil"

	"github.com/bukukasio/krm-functions/pkg/common/fnutils"
	esapi "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"github.com/facebookgo/subset"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	openapispec "k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	pgbouncerImage          = "gcr.io/beecash-prod/pgbouncer:bitnami-1.17.0-debian-11-r7"
	prometheusExporterImage = "spreaker/prometheus-pgbouncer-exporter"
	cpuLimit                = "500m"
	cpuRequest              = "10m"
	memoryLimit             = "500Mi"
	memoryRequest           = "50Mi"
)

// +kubebuilder:object:root=true
type FunctionConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:",inline"`
	Spec              spec `json:"spec"`
}

type spec struct {
	PartOf           string            `json:"part-of"`
	App              string            `json:"app"`
	ConnectionSecret string            `json:"connectionSecret"`
	Config           map[string]string `json:"config,omitempty"`
}

func (conf FunctionConfig) GetpgbouncerContainers() []corev1.Container {
	Containers := []corev1.Container{
		{
			Image: pgbouncerImage,
			EnvFrom: []corev1.EnvFromSource{
				{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: conf.Spec.ConnectionSecret,
						},
					},
				},
			},
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(cpuLimit),
					corev1.ResourceMemory: resource.MustParse(memoryLimit),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(cpuRequest),
					corev1.ResourceMemory: resource.MustParse(memoryRequest),
				},
			},
			Name: "pgbouncer",
			LivenessProbe: &corev1.Probe{
				InitialDelaySeconds: 60,
				PeriodSeconds:       10,
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString(intstr.FromInt(int(6432))),
					},
				},
			},
			ReadinessProbe: &corev1.Probe{
				InitialDelaySeconds: 20,
				PeriodSeconds:       10,
				FailureThreshold:    6,
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString(intstr.FromInt(int(6432))),
					},
				},
			},
			Lifecycle: &corev1.Lifecycle{
				PreStop: &corev1.LifecycleHandler{
					Exec: &corev1.ExecAction{
						Command: []string{
							"/bin/sh",
							"-c",
							"sleep 15 && psql $PGBOUNCER_DB_ADMIN_URL -c 'PAUSE $DB_NAME;'",
						},
					},
				},
			},
		},
		{
			Image: prometheusExporterImage,
			Name:  "prometheus-pgbouncer-exporter",
			Env: []corev1.EnvVar{
				{
					Name:  "PGBOUNCER_EXPORTER_HOST",
					Value: "0.0.0.0",
				},
				{
					Name:  "PGBOUNCER_USER",
					Value: "$(POSTGRESQL_USERNAME)",
				},
				{
					Name:  "PGBOUNCER_PASS",
					Value: "$(POSTGRESQL_PASSWORD)",
				},
				{
					Name:  "PGBOUNCER_PORT",
					Value: "6432",
				},
			},
			EnvFrom: []corev1.EnvFromSource{
				{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: conf.Spec.ConnectionSecret,
						},
					},
				},
			},
			Ports: []corev1.ContainerPort{
				{
					Name:          "pgb-metrics",
					ContainerPort: 9127,
				},
			},
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(cpuLimit),
					corev1.ResourceMemory: resource.MustParse(memoryLimit),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(cpuRequest),
					corev1.ResourceMemory: resource.MustParse(memoryRequest),
				},
			},
		},
	}
	return Containers
}

func GetTypeMeta(kind string, apiversion string) metav1.TypeMeta {
	typeMeta := metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiversion,
	}
	return typeMeta
}

func getName(conf spec) string {
	name := conf.PartOf + "-" + "pgbouncer"
	return name
}

func (conf FunctionConfig) GetObjectMeta() metav1.ObjectMeta {
	objectMeta := metav1.ObjectMeta{
		Name: getName(conf.Spec),
		Labels: map[string]string{
			"app":     getName(conf.Spec),
			"part-of": conf.Spec.PartOf,
		},
	}
	return objectMeta
}

func (conf FunctionConfig) GetMetaLabelSelector() *metav1.LabelSelector {
	metaLabelSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": getName(conf.Spec),
		},
	}
	return metaLabelSelector
}

func (conf FunctionConfig) getService() corev1.Service {
	service := corev1.Service{
		TypeMeta:   GetTypeMeta("Service", "v1"),
		ObjectMeta: conf.GetObjectMeta(),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       6432,
					Name:       "pgbouncer",
					TargetPort: intstr.IntOrString(intstr.FromInt(int(6432))),
					Protocol:   "TCP",
				},
			},
			Selector: conf.GetMetaLabelSelector().MatchLabels,
		},
	}
	return service
}

func (conf FunctionConfig) getDeployment() appsv1.Deployment {
	deployment := appsv1.Deployment{
		TypeMeta:   GetTypeMeta("Deployment", "apps/v1"),
		ObjectMeta: conf.GetObjectMeta(),
		Spec: appsv1.DeploymentSpec{
			Selector: conf.GetMetaLabelSelector(),
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{IntVal: 0},
					MaxSurge:       &intstr.IntOrString{IntVal: 2},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: conf.GetObjectMeta(),
				Spec: corev1.PodSpec{
					Containers: conf.GetpgbouncerContainers(),
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									TopologyKey: "kubernetes.io/hostname",
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app",
												Operator: metav1.LabelSelectorOpIn,
												Values: []string{
													getName(conf.Spec),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}

func (conf FunctionConfig) getPodMonitor() monitoringv1.PodMonitor {
	podMonitor := monitoringv1.PodMonitor{
		TypeMeta:   GetTypeMeta("PodMonitor", "monitoring.coreos.com/v1"),
		ObjectMeta: conf.GetObjectMeta(),
		Spec: monitoringv1.PodMonitorSpec{
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Path:        "/pgbouncer-metrics",
					Port:        "pgb-metrics",
					HonorLabels: true,
				},
			},
			Selector: *conf.GetMetaLabelSelector(),
		},
	}
	return podMonitor
}

func (conf FunctionConfig) getConfigMap() corev1.ConfigMap {
	configMap := corev1.ConfigMap{
		TypeMeta:   GetTypeMeta("ConfigMap", "v1"),
		ObjectMeta: conf.GetObjectMeta(),
		Data:       conf.Spec.Config,
	}
	return configMap
}

func (conf FunctionConfig) getPodDisruptionBudget() policyv1.PodDisruptionBudget {
	podDisruptionBudget := policyv1.PodDisruptionBudget{
		TypeMeta:   GetTypeMeta("PodDisruptionBudget", "policy/v1beta1"),
		ObjectMeta: conf.GetObjectMeta(),
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &intstr.IntOrString{IntVal: 1},
			Selector:     conf.GetMetaLabelSelector(),
		},
	}
	return podDisruptionBudget
}

func addConfigMapReference(d *appsv1.Deployment, cmName string) {
	configmapRef := []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cmName,
				},
			},
		},
	}
	for container := range d.Spec.Template.Spec.Containers {
		if d.Spec.Template.Spec.Containers[container].Name == "pgbouncer" {
			d.Spec.Template.Spec.Containers[container].EnvFrom = append(configmapRef, d.Spec.Template.Spec.Containers[container].EnvFrom...)
		}
	}
}

// Validation function for ExternalSecrets  -  validates that the secret contains all the required fields
func validateConnectionSecret(secret *esapi.ExternalSecret) error {
	// Fields that must be present in the secret
	expected_fields := []string{
		"POSTGRESQL_HOST",
		"POSTGRESQL_PORT",
		"POSTGRESQL_USERNAME",
		"POSTGRESQL_PASSWORD",
		"POSTGRESQL_DATABASE",
	}
	var data []string
	for _, s := range secret.Spec.Data {
		data = append(data, s.SecretKey)
	}
	issubset := subset.Check(expected_fields, data)
	if issubset {
		return nil
	} else {
		return fmt.Errorf("Some of the fields are missing from secret, Expected fields list %v", expected_fields)
	}
}

// KRMFunctionConfig.Filter is called from kio.Filter, which handles Results/errors appropriately
// errors break the pipeline, results are appended to the resource lists' Results
func (f *FunctionConfig) Filter(items []*kyaml.RNode) ([]*kyaml.RNode, error) {
	for _, item := range items {
		if item.GetKind() == "ExternalSecret" {
			targetName, err := item.GetString("spec.target.name")
			if err != nil {
				return nil, fmt.Errorf("Error parsing target name from ExternalSecret: %v", err)
			}
			if targetName == f.Spec.ConnectionSecret {
				itemExternalSecret, err := fnutils.ParseRNodeExternalSecret(item)
				if err != nil {
					return nil, fmt.Errorf("Error parsing ExternalSecret from Rnode: %v", err)
				}
				err = validateConnectionSecret(itemExternalSecret)
				if err != nil {
					return nil, fmt.Errorf("ConnectionSecret is invalid: %v", err)
				}
			}
		}
	}
	svc := f.getService()
	deployment := f.getDeployment()
	podmonitor := f.getPodMonitor()
	if f.Spec.Config != nil {
		cm := f.getConfigMap()
		cmRNode, _ := fnutils.MakeRNode(cm)
		items = append(items, cmRNode)
		addConfigMapReference(&deployment, cm.ObjectMeta.Name)
	}
	pdb := f.getPodDisruptionBudget()
	newNodes, err := fnutils.MakeRNodes(&svc, &deployment, &podmonitor, &pdb)
	items = append(items, newNodes...)
	return items, err
}

func (a FunctionConfig) Schema() (*openapispec.Schema, error) {
	var err error
	crdFile, err := ioutil.ReadFile("crd/pgbouncer/krm_functionconfigs.yaml")
	if err != nil {
		return nil, errors.WrapPrefixf(err, "\n reading crd file")
	}
	pgbouncerCrd := string(crdFile)
	schema, err := framework.SchemaFromFunctionDefinition(resid.NewGvk("krm", "pgbouncer", "FunctionConfig"), pgbouncerCrd)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "\n parsing pgbouncer crd")
	}
	return schema, errors.WrapPrefixf(err, "\n parsing pgbouncer schema")
}
