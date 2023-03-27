SHELL = /bin/zsh


test: check-function-var
	cat ${function}/example/resource_list.yaml | go run ${function} -

e2e: check-function-var build
	kustomize build --enable-alpha-plugins example/${function}

debug: check-function-var
	dlv debug ${function} -r <(cat ${function}/example/resource_list.yaml)

build: check-function-var
	docker build . --build-arg=FUNCTION=${function} -t gcr.io/beecash-prod/infra/krm-functions/${function}:latest

crd: check-function-var
	controller-gen crd paths=./pkg/workloads output:crd:dir=crd/workloads
	controller-gen crd paths=./pkg/pgbouncer output:crd:dir=crd/pgbouncer
	controller-gen crd paths=./pkg/pubsub output:crd:dir=crd/pubsub

check-function-var:
ifndef function
	# function variabl has to be for the directive
	$(error function . is undefined, run 'make <directive> function=<function-name>')
endif
