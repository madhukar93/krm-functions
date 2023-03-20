# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.19 AS builder
RUN go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.10.0
ENV CGO_ENABLED=0
ARG FUNCTION
WORKDIR /go/src/
# do the go mod stuff in a separate layer/image?
COPY go.mod .
COPY go.sum .
RUN go mod download 
COPY pkg/ pkg/
COPY cmd/${FUNCTION}/*.go .
RUN --mount=type=cache,target=/root/.cache/go-build go build -mod readonly -v -o /usr/local/bin/config-function ./ 

RUN controller-gen crd paths=./pkg/${FUNCTION} output:crd:dir=crd/${FUNCTION}

FROM alpine:3
ARG FUNCTION
COPY --from=builder /usr/local/bin/config-function /usr/local/bin/config-function
COPY --from=builder /go/src/crd/ crd/
COPY --from=builder /go/src/pkg/${FUNCTION}/*.yaml templates/${FUNCTION}/
CMD ["config-function"]
