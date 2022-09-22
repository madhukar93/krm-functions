package main

import (
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type functionConfig struct {
	TypeMeta   metav1.TypeMeta
	ObjectMeta metav1.ObjectMeta
	Spec       spec `yaml:"spec"`
}

type spec struct {
	PartOf     string            `yaml:"part-of"`
	App        string            `yaml:"app"`
	Connection connection        `yaml:"connection,omitempty"`
	Config     map[string]string `yaml:"config,omitempty"`
}

type connection struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Database          string `yaml:"database"`
	CredentialsSecret string `yaml:"credentialsSecret"`
}

func GetpgbouncerContainers() []corev1.Container {
	Containers := []corev1.Container{
		{
			Image: "gcr.io/beecash-prod/pgbouncer:1.14.working",
			Name:  "pgbouncer",
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
			Image: "spreaker/prometheus-pgbouncer-exporter",
			Name:  "prometheus-pgbouncer-exporter",
			Ports: []corev1.ContainerPort{
				{
					Name:          "pgb-metrics",
					ContainerPort: 9127,
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

func (conf functionConfig) GetObjectMeta() metav1.ObjectMeta {
	objectMeta := metav1.ObjectMeta{
		Name: conf.ObjectMeta.Name,
		Labels: map[string]string{
			"app":     conf.Spec.App,
			"part-of": conf.Spec.PartOf,
		},
	}
	return objectMeta
}

func (conf functionConfig) GetMetaLabelSelector() *metav1.LabelSelector {
	metaLabelSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app":     conf.Spec.App,
			"part-of": conf.Spec.PartOf,
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
			Selector: map[string]string{
				"app":     conf.Spec.App,
				"part-of": conf.Spec.PartOf,
			},
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
					Containers: GetpgbouncerContainers(),
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
													"pgbouncer",
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

func main() {
	var func_config functionConfig
	func_config.ObjectMeta.Name = "pgbouncer"
	func_config.Spec.App = "tokko"
	func_config.Spec.PartOf = "tokko-coupon"
	func_config.Spec.Config = map[string]string{
		"LISTEN_PORT":                 "6432",
		"MAX_CLIENT_CONN":             "1000",
		"PGBOUNCER_DEFAULT_POOL_SIZE": "200",
		"PGBOUNCER_MAX_CLIENT_CONN":   "5000",
		"POSTGRESQL_HOST":             " 10.48.0.2",
	}

	// call service generate func
	svc_obj := makeService(func_config)
	svc, _ := yaml.Marshal(svc_obj)
	fmt.Println(string(svc))
	fmt.Println("---\n")

	// call deployment generate func
	deployment_obj := makeDeployment(func_config)
	deployment, _ := yaml.Marshal(deployment_obj)
	fmt.Println(string(deployment))

	//call podmonitor generate func
	podmonitor_obj := makePodMonitor(func_config)
	podmonitor, _ := yaml.Marshal(podmonitor_obj)
	fmt.Println(string(podmonitor))

	// call config map generate func
	cm_obj := makeConfigMap(func_config)
	cm, _ := yaml.Marshal(cm_obj)
	fmt.Println(string(cm))

	// call pdb map generate func
	pdb_obj := makePodDisruptionBudget(func_config)
	pdb, _ := yaml.Marshal(pdb_obj)
	fmt.Println(string(pdb))

}
