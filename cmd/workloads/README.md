# Workloads

KRM functions for running jobs, crons and rollout objects.

```yaml
---
kind: LummoContainer
spec:
  # everything in corev1.Container
  base: <name of some LummoContainer>
  grpc:
    port: 5000
  http:
    port: 8000
  configs:
    - foobar-api
  secrets:
  - foobar-api

---
# fields that match as is will be passed as is
# special function fields will
kind: LummoDeployment
spec:
  part-of: foobar
  app: foobar-api
  containers: # want to make it as close to corev1.Container
    - name: foobar-api
  #   atleast one container with name == spec.app should be there
  #   this is the 'main' container, while others are sidecar
      image: foobar
      command: ["python", "server.py"]
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
    configs:
    - foobar-api
    secrets:
    - foobar-api
---
kind: LummoJob
spec:
  part-of: foobar
  name: daily-foo-job
  container:
    - foobar # use LummoContainer as is
    - base: foobar # overrides LummoContainer used as base
      command: ["python", "batch.py"]
    - name: whatever
      image: blah:latest
```