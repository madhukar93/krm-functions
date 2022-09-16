FROM golang:1.18
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY . .
# TODO: why use this hard-coded path?
RUN go build -o /home/molybdenum/kpt-network-resource

FROM alpine:3.15
COPY --from=0 /home/molybdenum/kpt-network-resource /kpt-network-resource
ENTRYPOINT ["/kpt-network-resource"]
