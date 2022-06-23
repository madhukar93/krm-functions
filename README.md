# tokko-networking-kpt

## Description

Manage tokko network resources with Kpt

## Usage

### Fetch the package
`kpt pkg get https://github.com/bukukasio/tokko-networking-kpt tokko-networking-kpt`

### View package content
`kpt pkg tree tokko-networking-kpt`


We used the following kpt file to configure the manifests

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: tokko-networking-kpt
  annotations:
    config.kubernetes.io/local-config: "true"
info:
  description: Set up Networking resources for tokko-k8s services
pipeline:
  mutators:
  - image: gcr.io/kpt-fn/apply-setters:unstable
    configPath: setters.yaml
  - image: gcr.io/kpt-fn/apply-replacements:unstable
    configPath: replacements.yaml
```

The setter values are provided through setter files and the replacement files replaces necessary values in the pipeline.

### Function Invocation

The functions in the pipeline are invoked using

`kpt fn render`

### Example Usage

Configure the variables in `setters.yaml` file, currently the variables that exist are:
 - service-name
 - service-port
 - target-port
 - rule

Run `kpt fn render` and check the substituted values in the respective `service.yaml` and `ingress.yaml`

### Results

Check the manifests to see the values being set and replaced
