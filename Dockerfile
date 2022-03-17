# syntax=docker/dockerfile:1.4
FROM --platform=$BUILDPLATFORM golang:1.18-bullseye as build

WORKDIR /workspace
COPY go.mod go.sum .
RUN go mod download
COPY --from= . .

ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
  go build -ldflags '-w -extldflags "-static"' \
  -o /usr/local/bin/envoystatic ./cmd/envoystatic

#FROM --platform=$TARGETPLATFORM gcr.io/distroless/static:nonroot as tooling
FROM --platform=$TARGETPLATFORM ubuntu:22.04 as tooling

COPY --from=build /usr/local/bin /usr/local/bin

# The source tree is not meant to be preserved
VOLUME /workspace

# ARG ENVOY_VERSION=v1.21.1
# FROM --platform=$TARGETPLATFORM envoyproxy/envoy:${ENVOY_VERSION} as envoy
FROM --platform=$TARGETPLATFORM envoyproxy/envoy:v1.21.1 as envoy

COPY bootstrap/* /etc/envoy/bootstrap/

RUN set -e; \
  mkdir /etc/envoy/rds; \
  ln -s /etc/envoy/bootstrap/route.yaml /etc/envoy/rds/route.yaml

USER envoy:nogroup

CMD [ "envoy", \
  "-c", "/etc/envoy/bootstrap/envoy.yaml", \
  "--service-cluster", "envoystatic", \
  "--service-node", "envoystatic", \
  "-l", "info" ]
