package main

import (
	"bytes"
	"fmt"

	"github.com/bukukasio/krm-functions/pkg/fnutils"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

type functionConfig struct {
	TypeMeta   metav1.TypeMeta
	ObjectMeta metav1.ObjectMeta
	Spec       spec `yaml:"spec"`
}

type spec struct {
	PartOf     string            `json:"part-of"`
	App        string            `json:"app"`
	Connection connection        `json:"connection,omitempty"`
	Config     map[string]string `json:"config,omitempty"`
}

type connection struct {
	Host              string `json:"host"`
	Port              string `json:"port"`
	Database          string `json:"database"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	CredentialsSecret string `json:"credentialsSecret"`
}

func (conf functionConfig) GetpgbouncerContainers() []corev1.Container {
	Containers := []corev1.Container{
		{
			Image: pgbouncerImage,
			EnvFrom: []corev1.EnvFromSource{
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: getName(conf.Spec),
						},
					},
				},
				{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: conf.Spec.Connection.CredentialsSecret,
						},
					},
				},
			},
			Env: []corev1.EnvVar{
				{
					Name: "SERVICE_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.labels['app']",
						},
					},
				},
				{
					Name:  "POSTGRESQL_USERNAME",
					Value: conf.Spec.Connection.Username,
				},
				{
					Name:  "POSTGRESQL_PASSWORD",
					Value: conf.Spec.Connection.Password,
				},
				{
					Name:  "POSTGRESQL_DATABASE",
					Value: conf.Spec.Connection.Database,
				},
				{
					Name:  "POSTGRESQL_PORT",
					Value: conf.Spec.Connection.Port,
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
					Name:  "PGBOUNCER_USER",
					Value: conf.Spec.Connection.Username,
				},
				{
					Name:  "PGBOUNCER_PASS",
					Value: conf.Spec.Connection.Password,
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

func (conf functionConfig) GetConfigMapData() map[string]string {
	configMap := map[string]string{}

	for key, val := range conf.Spec.Config {
		configMap[key] = val
	}
	return configMap
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

func (conf functionConfig) GetObjectMeta() metav1.ObjectMeta {
	objectMeta := metav1.ObjectMeta{
		Name: getName(conf.Spec),
		Labels: map[string]string{
			"app":     getName(conf.Spec),
			"part-of": conf.Spec.PartOf,
		},
	}
	return objectMeta
}

func (conf functionConfig) GetMetaLabelSelector() *metav1.LabelSelector {
	metaLabelSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": getName(conf.Spec),
		},
	}
	return metaLabelSelector
}

func makeService(conf functionConfig) corev1.Service {
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

func makeDeployment(conf functionConfig) appsv1.Deployment {
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

func makePodMonitor(conf functionConfig) monitoringv1.PodMonitor {
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

func makeConfigMap(conf functionConfig) corev1.ConfigMap {
	configMap := corev1.ConfigMap{
		TypeMeta:   GetTypeMeta("ConfigMap", "v1"),
		ObjectMeta: conf.GetObjectMeta(),
		Data:       conf.GetConfigMapData(),
	}
	return configMap
}

func makePodDisruptionBudget(conf functionConfig) policyv1.PodDisruptionBudget {
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

func appendRnodes(in []*kyaml.RNode, objects ...any) ([]*kyaml.RNode, error) {
	out := []*kyaml.RNode{}
	var errors []error
	for _, o := range objects {
		r, err := fnutils.MakeRNode(o)
		if err != nil {
			errors = append(errors, err)
		}
		out = append(out, r)
	}
	if len(errors) > 0 {
		var combinedErrors bytes.Buffer
		for _, e := range errors {
			combinedErrors.WriteString(fmt.Sprintln("%s", e.Error()))
		}
		return nil, fmt.Errorf(combinedErrors.String())
	}
	return out, nil
}

func (f functionConfig) filter(in []*kyaml.RNode) ([]*kyaml.RNode, error) {
	svc := makeService(f)
	deployment := makeDeployment(f)
	podmonitor := makePodMonitor(f)
	cm := makeConfigMap(f)
	pdb := makePodDisruptionBudget(f)
	return appendRnodes(in, svc, deployment, podmonitor, cm, pdb)
}
