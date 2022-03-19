# syntax=docker/dockerfile:1.4
FROM --platform=$BUILDPLATFORM golang:1.18-bullseye as build

WORKDIR /workspace
COPY go.mod go.sum .
RUN go mod download
COPY --from= . .
RUN go test ./...

ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
  go build -ldflags '-w -extldflags "-static"' \
  -o /usr/local/bin/envoystatic ./cmd/envoystatic

FROM --platform=$TARGETPLATFORM gcr.io/distroless/static:nonroot as tooling

COPY --from=build /usr/local/bin /usr/local/bin

# The source tree is not meant to be preserved
VOLUME /workspace

# Entrypoint is unuseful for the recommended workflow (as FROM at build time)
# but helpful to test pipelines using local docker
ENTRYPOINT [ "/usr/local/bin/envoystatic" ]

# Envoy distroless is not yet multi-arch, and build arg as from tag didn't work
# ARG ENVOY_VERSION=v1.21.1
# FROM --platform=$TARGETPLATFORM envoyproxy/envoy:${ENVOY_VERSION} as envoy
FROM --platform=$TARGETPLATFORM envoyproxy/envoy:v1.21.1 as envoy

COPY bootstrap/* /etc/envoy/bootstrap/

RUN set -e; \
  mkdir /etc/envoy/rds; \
  ln -s /etc/envoy/bootstrap/route.yaml /etc/envoy/rds/route.yaml

USER envoy:nogroup

EXPOSE 8080/tcp

CMD [ "envoy", \
  "-c", "/etc/envoy/bootstrap/envoy.yaml", \
  "--service-cluster", "envoystatic", \
  "--service-node", "envoystatic", \
  "-l", "info" ]
