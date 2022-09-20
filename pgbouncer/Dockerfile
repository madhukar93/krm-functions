
FROM golang:1.18-stretch
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY . .
RUN go mod download
RUN go build -v -o /usr/local/bin/function ./

FROM alpine:latest
COPY --from=0 /usr/local/bin/function /usr/local/bin/function
CMD ["function"]