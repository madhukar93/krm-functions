package main

import (
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/parser"
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

func main() {
	func_config := new(functionConfig)

	// create the template
	fn := framework.TemplateProcessor{
		// Templates input
		TemplateData: func_config,
		ResourceTemplates: []framework.ResourceTemplate{
			{
				Templates: parser.TemplateStrings(serviceTemplate + "---\n" + deploymentTemplate + "---\n" + podMonitorTemplate),
			},
		},
	}

	cmd := command.Build(fn, command.StandaloneDisabled, false)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// TODO: write functions to generate deployment, service and pod monitor spec

var serviceTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: {{ .Spec.App }}
spec:
  selector:
    app: pgbouncer
  ports:
  - name: pgbouncer
    port: 6432
    targetPort: 6432
    protocol: TCP
`

var deploymentTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Spec.App }}
  labels:
    app: pgbouncer
    component: pgbouncer
spec:
  selector:
    matchLabels:
      app:  pgbouncer
  replicas: 2
  strategy:
    rollingUpdate:
      maxSurge: 2
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: pgbouncer
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - pgbouncer
            topologyKey: kubernetes.io/hostname
      containers:
      - image: gcr.io/beecash-prod/pgbouncer:1.14.working
        name: pgbouncer
        env:
          - name: DATABASE_URL
            value: "postgres://$(DB_USER):$(DB_PASSWORD)@$(DATABASE_HOST):5432/$(DB_NAME)"
          - name: ADMIN_USERS
            value: $(DB_USER)
        livenessProbe:
          tcpSocket:
            port: 6432
          initialDelaySeconds: 60
          periodSeconds: 10
        readinessProbe:
          tcpSocket:
            port: 6432
          initialDelaySeconds: 20
          failureThreshold: 6
          periodSeconds: 10
        lifecycle:
          preStop:
            exec:
              # Remove pod from service but keep it active for some time
              command: ['/bin/sh', '-c', 'sleep 15 && psql $PGBOUNCER_DB_ADMIN_URL -c "PAUSE $DB_NAME;"']
        resources:
          requests:
            cpu: 50m
            memory: 100Mi
          limits:
            cpu: 1
            memory: 500Mi
      - image: spreaker/prometheus-pgbouncer-exporter
        name: prometheus-pgbouncer-exporter
        env:
          - name: PGBOUNCER_PASS
            valueFrom:
              secretKeyRef:
                name: tokko-api
                key: DB_PASSWORD
          - name: PGBOUNCER_EXPORTER_HOST
            value: 0.0.0.0
          - name: PGBOUNCER_PORT
            value: '6432'
        ports:
          - name: pgb-metrics
            containerPort: 9127
        resources:
          requests:
            cpu: 10m
            memory: 50Mi
          limits:
            cpu: 100m
            memory: 500Mi
`

var podMonitorTemplate = `apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: {{ .Spec.App }}
spec:
  podMetricsEndpoints:
  - path: /pgbouncer-metrics
    port: pgb-metrics
    honorLabels: true
  selector:
    matchLabels:
      app: pgbouncer
`
