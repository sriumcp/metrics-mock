FROM golang:1.15-buster as builder

WORKDIR /metrics-mock

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go
COPY prometheus.go prometheus.go
COPY newrelic.go newrelic.go

EXPOSE 8080

# ARG config_url
ENV CONFIG_URL "https://raw.githubusercontent.com/iter8-tools/metrics-mock/main/testdata/uriconfs.yaml"

RUN go build
CMD ["./metrics-mock"]