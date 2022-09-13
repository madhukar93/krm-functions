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
part-of: foobar
app: foobar-api
domains:
- a.test.com
- b.test.com
grpc: true
routes:
- match: Path(`/hello`)
  vpn: true
- match: Path(`/world`)
```

## Workloads

runs your containers for servers and jobs

```yaml
---
kind: Deployment
part-of: foobar
app: foobar-api
container:
  command: ["python", "server.py"]
  image: foobar
  port: 80
```

## Design priciples

1. Hermetic builds - will not talk to network etc, will just consist of declarative config - this allows us do validate config earlier
   (shift left!).
2. Should contain minimal (next to none k8s specific config). Should contain general terms
3. Should build upon our current deployment tooling and code (work with argocd, argo rollouts etc) and seek to replace
   it for considerable gains only. The current implementation has the weakness that it is not sufficiently abstract but
   its strengths are that it is built on solid foundations. Let's not fix what's not broken.
4. We will try to server the 90% usecase first, it can be 10% leaky. KISS - for development and usage. No over-abstraction.
   This approach doesn't force you to stick to it, it's just kustomize after all.

## Roadmap

- [ ] networking resources
- [ ] workloads
- [ ] autoscaling
- [ ] postgres

## FAQ

### 1. why not OAM

OAM will make us code to its specification, which will be more complex than what we can come up with since it has to be
more extensible and we can do something simpler. It will also make us adhere to its semantics which we will have to learn
and we already have some implementations in mind.

There would have been benefits to an out of the box implementation if we had out of the box implementations we could use
but the standard and implementations are at a nascent stage. Kubevela, it's canonical implementation renders cue templates to generate
resources on the server side which is very different from the implementation we are going for. Plus this doesn't look very appetizing
<https://kubevela.io/docs/tutorials/k8s-object#deploy-with-cli>

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
