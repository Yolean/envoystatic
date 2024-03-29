# syntax=docker/dockerfile:1.4
ARG envoy_version="v1.25.1"

FROM --platform=$BUILDPLATFORM golang:1.20-bullseye as build

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

FROM --platform=$TARGETPLATFORM envoyproxy/envoy:${envoy_version} as envoy

# This layer is used as the -debug image, until we have a skaffold verify workflow we need curl
RUN set -ex; export DEBIAN_FRONTEND=noninteractive; runDeps='curl'; \
  apt-get update && apt-get install -y $runDeps --no-install-recommends; \
  rm -rf /var/lib/apt/lists; \
  rm -rf /var/log/dpkg.log /var/log/alternatives.log /var/log/apt /root/.gnupg

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

FROM --platform=$TARGETPLATFORM envoyproxy/envoy-distroless:${envoy_version} as envoy-distroless

COPY --from=envoy /etc/passwd /etc/group /etc/

COPY --from=envoy /etc/envoy /etc/envoy

USER envoy:nogroup

EXPOSE 8080/tcp

CMD [ \
  "-c", "/etc/envoy/bootstrap/envoy.yaml", \
  "--service-cluster", "envoystatic", \
  "--service-node", "envoystatic", \
  "-l", "info" ]
