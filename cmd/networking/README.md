# Inject Routes

Function to generate network resources

## Inputs and Outputs

```yaml
---
# input
kind: LummoNetworking
metadata:
  name: test
spec:
  app: test-server
  grpc: true # new ingressroute with only grpc stuff - always internal, uses some fake domain, uses web entrypoint, scheme: h2c
  domains: # creates certs, host matching rules
  - a.test.com
  - b.test.com
  routes: # routes to service for the port named https
  # ingressroute entrypoint is websecure
  - match: Path(`/hello`) # these will always be https
    vpn: true
  - match: Path(`/world`)
---
# output
kind: Certificate
dnsNames:
  - a.test.com
  - b.test.com
---
kind: IngressRoute
metadata:
  name: test-server
spec:
  entryPoints:
  - web
  routes:
  - match: (Host(`a.test.com`) || Host(`a.test.com`)) && Path(`/hello`)
    kind: Rule
    services:
    - name: echo-server
      port: 80
  - match: (Host(`b.test.com`) || Host(`a.test.com`)) && Path(`/world`)
    kind: Rule
    services:
    - name: echo-server
      port: 80
  tls:
    secretName: test-server-cert
---
Kind: Service
metadata:
  name: echo-server
  ...
```

## TODO

- [x] deduce service from app (app label matches, app key in fn config)

```
a deployment that doesn’t have an app label is invalid - to be eventually validated using something like kubeval during CI, or admission control during apply time.
```

- [x] if the function doesn’t find this app (deployment/rollout with app label) exit with error

- [x] when you find a deployment, use container port and app label in pod template.spec to create a service

- [x] add vpn flag

- [x] support grpc
