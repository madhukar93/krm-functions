build:
	docker build . -t networking:latest

e2e: build
	cd demo/config && kustomize build --enable-alpha-plugins

.PHONY: build
