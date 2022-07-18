FROM golang:1.17-alpine3.15
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /home/molybdenum/kpt-network-resource

FROM alpine:3.15
COPY --from=0 /home/molybdenum/kpt-network-resource /kpt-network-resource
ENTRYPOINT ["/kpt-network-resource"]
