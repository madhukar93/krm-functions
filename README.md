# krm-functions

This repository will contain various krm-functions that produce k8s configuration for deploying applications and jobs

## usage

- lummo-pay
- tokko-api

## Functions

---

### Networking

sets up k8s network resources like ingress rules, ssl certs,
load balancers needed to make services talk to each other and
the outside world.

```yaml
kind: LummoNetworking
metadata:
  name: test
spec:
  app: test-server
  grpc: true
  domains:
  - a.test.com
  - b.test.com
  routes:
  - match: Path(`/hello`) # these will always be https
    vpn: true
  - match: Path(`/world`)
```

## Workloads

runs your containers for servers and jobs

```yaml
---
kind: LummoDeployment
spec:
  part-of: foobar
  app: foobar-api
  containers:
    command: ["python", "server.py"]
    image: foobar
    ports: ...
    configs:
      - "tokko-api"
    secrets:
      - "tokko-api" # contains DB connection details also, which should match with pgbouncer
    resources: # make this required
  monitoring: # it will just add DD envs vars
    datadog: true
    prometheus: # not needed for now
      endpoint: '/metrics'
      port: 1234 # when sidecar
  strategy: {} # as is, produce rollout
  scaling: # always use keda
    minreplica: 1
    maxreplica: 10
    cpu:
      target: 50
    memory:
      target: 80 # (of requests)
    pubsubTopic:
      - name: some-queue
        size: 10000 # TODO: allow higher level controls like latency and throughput
    # no prom stuff for now
```

## pubsub

```yaml
---
kind: LummoTopic
spec:
  prefix: dev-
  topics:
  - topicA
  - topicB
---
kind: LummoSubscription
spec:
  prefix: dev-
  config:
    ackDeadlineSeconds: 10
    maxDeliveryAttempts: 5
    ttl: 2678400s
    messageRetentionDuration: 604800s
    maximumBackoff: 600s
    minimumBackoff: 300s
  subscriptions:
  - topic: topicA
    subscription: subA
  - topic: topicB
    subscription: subB
```
## pgbouncer

```yaml
apiVersion: LummoKRM
kind: pgbouncer
metadata:
  name: tokko-api-pgbouncer
spec:
  app: foobar-api
  part-of: foobar
  spec:
    connectionSecret: tokko-api-postgres-creds # is of ConnectionSecret type which has the fields <TODO>
    config: # creates config map
      POOL_SIZE: 100
      # etc
```

vault infra/postgres/tokko-api-postgres/creds

```
# kube/cloudsql/tokko-api/postgres
- kustomization.yaml
- db-secret.yaml # we can identify and map it by metadata
```

## Design priciples

1. Hermetic builds - will not talk to network etc, will just consist of declarative config - this allows us do validate config earlier
   (shift left!).
2. Should contain minimal (next to none k8s specific config). Should contain general terms
3. Should build upon our current deployment tooling and code (work with argocd, argo rollouts etc) and seek to replace
   it for considerable gains only. The current implementation has the weakness that it is not sufficiently abstract but
   its strengths are that it is built on solid foundations. Let's not fix what's not broken.
4. We will try to serve the 90% usecase first, it can be 10% leaky. KISS - for development and usage. No over-abstraction.
   This approach doesn't force you to stick to it, it's just kustomize after all, and there are easy escape hatches and alternatives.

## Roadmap

- [x] networking resources
- [x] workloads
- [x] autoscaling
- [x] monitoring
- [x] canary
- [x] pgbouncer
- [ ] environments
- [ ] container template
- [x] argocd integration
- [x] reloader
- [ ] vault integration
- [ ] pubsub

### supporting multiple environment

For now we will just use kustomize overlays to support multiple environments. The functions will live in the base and the overlays will just contain the environment specific config.

### container template

The same container is used in multiple workloads. We can use a container template to define it once and reuse it in multiple workloads. LummoContainer can extend other LummoContainers for workload specific config.

eg. command for different jobs will be different but the image and configmaps, secrets will be the same.

probes will be very specific to the workload, so we will likely need to define them in the base container.

### argocd integration

Argocd has to run KRM functions. We can run KRM functions as libraries or containers. The containerized approach will require dind. We can use the kustomize plugin approach to run the functions as libraries.

## FAQ

### 1. why not OAM

OAM will make us code to its specification, which will be more complex than what we can come up with since it has to be
more extensible and we can do something simpler. It will also make us adhere to its semantics which we will have to learn
and we already have some implementations in mind.

There would have been benefits to an out of the box implementation if we had out of the box implementations we could use
but the standard and implementations are at a nascent stage. Kubevela, it's canonical implementation renders cue templates to generate
resources on the server side which is very different from the implementation we are going for. Plus this doesn't look very appetizing
<https://kubevela.io/docs/tutorials/k8s-object#deploy-with-cli>

### Why not helm?

templating is bad

- templating is just programming in an inferior 'stringly typed' environment
- coding, debugging, testing etc is harder
- Writing significant logic is hard

 why KRM functions are better -

- can use any programming language and it's ecosystem
- functions have single responsibility
- functions are composable (they can be 'piped'), easier to do cross cutting concerns
- can reuse code between k8s operators and client side functions
- can be used for more than just config generation, can be used for transformation and validation

## links

### project

- RFC <https://bukukas.atlassian.net/wiki/spaces/TD/pages/521371768/Creating+kustomize+kpt+extension+to+manage+network+resources>.
- issue tracker

### projects to watch

- <https://oam.dev/> <https://github.com/oam-dev/spec>
- <https://github.com/GoogleContainerTools/kpt>
- <https://github.com/GoogleContainerTools/kpt-backstage-plugins>
- <https://crossplane.io/>
- <https://kubevela.io/>
- <https://docs.airshipit.org/>
- <https://cloud.google.com/config-connector/docs/overview>

### Learn

- <https://www.youtube.com/watch?v=YlFUv4F5PYc>
- <https://kubectl.docs.kubernetes.io/guides/extending_kustomize/>
- <https://www.gitops.tech/>
- <https://github.com/kubernetes/design-proposals-archive/blob/main/architecture/declarative-application-management.md>
- <https://cloud.google.com/blog/products/containers-kubernetes/understanding-configuration-as-data-in-kubernetes>
- <https://cloud.google.com/blog/topics/developers-practitioners/build-platform-krm-part-1-whats-platform>
- <https://pkg.go.dev/sigs.k8s.io/kustomize/kyaml/fn/framework>
- <https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md>
- <https://kpt.dev/book/02-concepts/03-functions>

### packages

- yaml & KRM
  - sigs.k8s.io/kustomize/kyaml/kio
  - k8s.io/apimachinery/pkg/util/yaml
  - sigs.k8s.io/yaml
  - sigs.k8s.io/kustomize/kyaml/yaml
- k8s API
  - k8s.io/api/apps/v1
  - k8s.io/api/core/v1
  - k8s.io/apimachinery/pkg/apis/meta/v1
