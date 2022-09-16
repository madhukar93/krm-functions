SHELL = /bin/zsh

test:
	cat example/resource_list.yaml | go run . -

debug:
	dlv debug . -r <(cat example/resource_list.yaml)

build:
	docker build . -t gcr.io/beecash-prod/infra/krm-functions/workloads:latest

.PHONY: build
