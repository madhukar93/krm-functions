package main

import (
	"testing"

	fntesting "github.com/bukukasio/krm-functions/pkg/testing"
)

var deploymentInput = `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items: 
- apiVersion: LummoKRM
  kind: LummoDeployment
  metadata:
    name: lummo-app
  spec:
    part-of: foobar
    app: foobar-api
    containers:
    - name: foobar-api
      image: foobar
      command: ["python", "server.py"]
      http:
        port: 2000
      secrets:
      - foobar-api-database-secrets
      configs:
      - foobar-api-config
    scaling:
      enabled: true 
      minreplica: 1
      maxreplica: 10
      cpu:
        target: "60"
      memory:
        target: "80"
      pubsubTopic:
        name:  dev-tokko-subscription.catalog-integration-product-added
        size: "500"
`

var deploymentOutput = `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    creationTimestamp: null
    labels:
      app: foobar-api
      part-of: foobar
    name: foobar-api
  spec:
    selector:
      matchLabels:
        app: foobar-api
        part-of: foobar
    strategy: {}
    template:
      metadata:
        creationTimestamp: null
        labels:
          app: foobar-api
          part-of: foobar
      spec:
        containers:
        - command:
          - python
          - server.py
          envFrom:
          - configMapRef:
              name: foobar-api-config
          - secretRef:
              name: foobar-api-database-secrets
          image: foobar
          name: foobar-api
          ports:
          - containerPort: 2000
            name: http
            protocol: TCP
          resources: {}
  status: {}
- apiVersion: v1
  kind: Service
  metadata:
    creationTimestamp: null
    labels:
      app: foobar-api
      part-of: foobar
    name: foobar-api
  spec:
    ports:
    - name: http
      port: 2000
      targetPort: 2000
    selector:
      app: foobar-api
      part-of: foobar
  status:
    loadBalancer: {}
- apiVersion: keda.sh/v1alpha1
  kind: ScaledObject
  metadata:
    creationTimestamp: null
    labels:
      app: foobar-api
      part-of: foobar
    name: foobar-api
  spec:
    maxReplicaCount: 10
    minReplicaCount: 1
    scaleTargetRef:
      apiVersion: argoproj.io/v1alpha1
      kind: Rollout
      name: foobar-api
    triggers:
    - authenticationRef:
        name: keda-trigger-auth-gcp-credentials
      metadata:
        subscriptionName: dev-tokko-subscription.catalog-integration-product-added
        subscriptionSize: "500"
      type: gcp-pubsub
    - metadata:
        type: Utilization
        value: "60"
      type: memory
    - metadata:
        type: Utilization
        value: "60"
      type: cpu
  status: {}
`

var jobInput = `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items: 
- apiVersion: lummoKRM/v1
  kind: LummoJob
  metadata:
    name: test-job
  spec:
    part-of: foobar
    app: foobar-api
    containers:
    - name: foobar-api
      image: foobar
      command: ["python", "server.py"]
      secrets:
        - foobar-api-database-secrets
      configs:
        - foobar-api-config`

var jobOutput = `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: batch/v1
  kind: Job
  metadata:
    creationTimestamp: null
    labels:
      app: foobar-api
      part-of: foobar
    name: foobar-api
  spec:
    template:
      metadata:
        creationTimestamp: null
        labels:
          app: foobar-api
          part-of: foobar
        name: foobar-api
      spec:
        containers:
        - command:
          - python
          - server.py
          envFrom:
          - configMapRef:
              name: foobar-api-config
          - secretRef:
              name: foobar-api-database-secrets
          image: foobar
          name: foobar-api
          resources: {}
  status: {}
`

var cronInput = `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items: 
- apiVersion: lummoKRM/v1
  kind: LummoCron
  metadata:
    name: test-cron
  spec:
    part-of: foobar
    app: foobar-api
    schedule: "* * * * *"
    containers:
      - name: foobar-api
        image: foobar
        command: ["python", "server.py"]
        secrets:
        - foobar-api-database-secrets
        configs:
        - foobar-api-config`

var cronOutput = `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: batch/v1
  kind: CronJob
  metadata:
    creationTimestamp: null
    labels:
      app: foobar-api
      part-of: foobar
    name: foobar-api
  spec:
    jobTemplate:
      metadata:
        creationTimestamp: null
        labels:
          app: foobar-api
          part-of: foobar
        name: foobar-api
      spec:
        template:
          metadata:
            creationTimestamp: null
            labels:
              app: foobar-api
              part-of: foobar
            name: foobar-api
          spec:
            containers:
            - command:
              - python
              - server.py
              envFrom:
              - configMapRef:
                  name: foobar-api-config
              - secretRef:
                  name: foobar-api-database-secrets
              image: foobar
              name: foobar-api
              resources: {}
    schedule: '* * * * *'
  status: {}
`

func TestDeployment(t *testing.T) {
	if fntesting.Compare(appFunc, deploymentInput, deploymentOutput) != true {
		t.Fatal()
	}
}

func TestJobs(t *testing.T) {
	if fntesting.Compare(appFunc, jobInput, jobOutput) != true {
		t.Fatal()
	}
}

func TestCronJobs(t *testing.T) {
	if fntesting.Compare(appFunc, cronInput, cronOutput) != true {
		t.Fatal()
	}
}
