package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"testing"
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

var jobInput = ``
var jobOutput = ``

func compare(in string, expected_out string) bool {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		if err != nil {
			outC <- ""
		}
		outC <- buf.String()
	}()

	tmpfile, err := os.CreateTemp("", "test-input")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // noerrcheck
	if _, err := tmpfile.Write([]byte(in)); err != nil {
		log.Fatal(err)
	}
	if _, err := tmpfile.Seek(0, 0); err != nil {
		log.Fatal(err)
	}
	os.Stdin = tmpfile

	err = appFunc()
	if err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	if err := w.Close(); err != nil {
		log.Fatal(err)
	}

	out := <-outC
	if out != expected_out {
		log.Printf("expected %s\nbut got %s\n", expected_out, out)
		return false
	}
	return true
}

func TestDeployment(t *testing.T) {
	if compare(deploymentInput, deploymentOutput) != true {
		t.Fatal()
	}
}

func TestJobs(t *testing.T) {
	if compare(jobInput, jobOutput) != true {
		t.Fatal()
	}
}
