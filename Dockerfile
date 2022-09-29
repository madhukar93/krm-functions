# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.19 AS builder
ENV CGO_ENABLED=0
ARG FUNCTION_DIR
WORKDIR /go/src/
# do the go mod stuff in a separate layer/image?
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY pkg/ pkg/
COPY ${FUNCTION_DIR}/*.go .
RUN go build -v -o /usr/local/bin/config-function ./

FROM alpine:latest
COPY --from=builder /usr/local/bin/config-function /usr/local/bin/config-function
CMD ["config-function"]
