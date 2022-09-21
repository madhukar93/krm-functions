package main

import (
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type functionConfig struct {
	TypeMeta   metav1.TypeMeta
	ObjectMeta metav1.ObjectMeta
	Spec       spec `yaml:"spec"`
}

type spec struct {
	PartOf     string     `yaml:"part-of"`
	App        string     `yaml:"app"`
	Connection connection `yaml:"connection,omitempty"`
	Config     config     `yaml:"config,omitempty"`
}

type connection struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Database          string `yaml:"database"`
	CredentialsSecret string `yaml:"credentialsSecret"`
}

type config struct {
	// TODO
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

func makeService(conf functionConfig) corev1.Service {
	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: conf.ObjectMeta.Name,
		},
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
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: conf.ObjectMeta.Name,
			Labels: map[string]string{
				"app":     conf.Spec.App,
				"part-of": conf.Spec.PartOf,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     conf.Spec.App,
					"part-of": conf.Spec.PartOf,
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{IntVal: 0},
					MaxSurge:       &intstr.IntOrString{IntVal: 2},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     conf.Spec.App,
						"part-of": conf.Spec.PartOf,
					},
				},
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
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodMonitor",
			APIVersion: "monitoring.coreos.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: conf.ObjectMeta.Name,
		},
		Spec: monitoringv1.PodMonitorSpec{
			PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
				{
					Path:        "/pgbouncer-metrics",
					Port:        "pgb-metrics",
					HonorLabels: true,
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     conf.Spec.App,
					"part-of": conf.Spec.PartOf,
				},
			},
		},
	}
	return podMonitor
}

func main() {
	var func_config functionConfig
	func_config.ObjectMeta.Name = "testsvc"
	func_config.Spec.App = "testapp"
	func_config.Spec.PartOf = "testapp-partof"

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

}
