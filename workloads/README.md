# Workloads

KRM functions for running jobs, crons and rollout objects.

```yaml
---
# fields that match as is will be passed as is
# special function fields will
kind: LummoDeployment
spec:
  part-of: foobar
  app: foobar-api
  metadata:
    annotations: []
    labels: []
  containers: # want to make it as close to corev1.Container
    name: foobar-api
  # atleast one container with name == spec.app should be there
  # this is the 'main' container, while others are sidecar
    image: foobar
    command: ["python", "server.py"]
    # port: 80 nah, do the same stuff
  # all env vars are imported as is using envFrom
    grpc:
      port: 5000
    http:
      port: 8000
    configMaps:
      - foobar-api
    secrets:
    - foobar-api
    # same as k8s containers
    strategy: {} # TODO Lift from Rollout kind?
    scaling: {} # TODO: build KEDA resource
---
kind: LummoCron
spec:
  part-of: foobar
  name: foobar
  schedule: * * * */10
  container:
    command: ["python", "cron.py"]
    image: test-server-job
    configMaps:
    - foobar-api
    secrets:
    - foobar-api
---
kind: LummoJob
spec:
  part-of: foobar
  name: daily-foo-job
  container:
    command: ["python", "job.py"]
    image: test-server-job
    configMaps:
    - foobar-api
    secrets:
    - foobar-api
```
