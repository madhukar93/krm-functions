# Workloads

KRM functions for running jobs, crons and rollout objects.

```yaml
---
# fields that match as is will be passed as is
# special function fields will
kind: LummoApp
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
    networking:
      domains:
      - foobar.com
      grpc:
        port: 5000
      http:
        port: 8000
        probe:
          path: /health
          # port: 8000 automatic
          # scheme: http automatic
          # below can have defaults, already do, but can be overridden
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 3
        routes:
          - match: /api/v1/foobar
          - match: internal/dashboard
            vpn: true
  # all env vars are imported as is using envFrom
    configs:
      - foobar-api
    secrets:
      - foobar-api
    # same as k8s containers
    monitoring: {}
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
---
kind: LummoJob
spec:
  part-of: foobar
  name: daily-foo-job
  container:
    command: ["python", "job.py"]
    image: test-server-job
```
