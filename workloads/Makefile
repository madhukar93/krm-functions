SHELL = /bin/zsh

test:
	cat example/resource_list.yaml | go run . -

e2e: build
	cd example && kustomize build --enable-alpha-plugins .

debug:
	dlv debug . -r <(cat example/resource_list.yaml)

build:
	docker build . -t gcr.io/beecash-prod/infra/krm-functions/workloads:latest
