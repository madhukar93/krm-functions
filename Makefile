SHELL = /bin/zsh


test: check-function-var
	cat ${function}/example/resource_list.yaml | go run ${function} -

e2e: check-function-var build
	kustomize build --enable-alpha-plugins ${function}/example

debug: check-function-var
	dlv debug ${function} -r <(cat ${function}/example/resource_list.yaml)

debug-linux: check-function-var
	dlv debug ./${function} --check-go-version=false -r <(cat ${function}/example/resource_list.yaml)

build: check-function-var
	docker build . --build-arg=FUNCTION_DIR=${function} -t gcr.io/beecash-prod/infra/krm-functions/${function}:latest

check-function-var:
ifndef function
	# function variabl has to be for the directive
	$(error function . is undefined, run 'make <directive> function=<function-name>')
endif
