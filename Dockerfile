FROM golang:1.15-buster as builder

WORKDIR /metrics-gen

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go
COPY prometheus.go prometheus.go
COPY newrelic.go newrelic.go

EXPOSE 8080

# ARG config_url
ENV CONFIG_URL "https://gist.githubusercontent.com/sushmarchandran/f0b51ea57642a96dc4269ec417df45db/raw/4ecc514db28cf434102a65abc1744f5ec5c46e15/uriconfs.yaml"

RUN go build
CMD ["./metrics-gen"]