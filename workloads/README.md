# Workloads

KRM functions for running jobs, crons and rollout objects.

```yaml
---
# fields that match as is will be passed as is
# special function fields will
kind: Deployment
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
    configMaps:
      - foobar-api
    secrets:
    - foobar-api
    # same as k8s containers
    env:
      -
      probes:
        readiness:
        liveness:
    strategy:
      ... # exactly the same as rollout?
---
kind: Cron
part-of: foobar
name: foobar
schedule: * * * */10
container:
  command: ["python", "cron.py"]
  image: test-server-job
---
kind: Job
part-of: foobar
name: daily-foo-job
container:
  command: ["python", "job.py"]
  image: test-server-job
```
