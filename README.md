# envoy-static-webserver

Rethinking the static web server for immutable infrastructure.

Container images  has a build step. The old static file server requirements no longer apply.
We can instead pre-process the directory structure to be hosted.

While other static file servers mount a volume,
here we build a container that represents the content at a specific state.
There is no longer a need to check the underlying file system for changes.

Caching can be the task of the underlying file system.
With Envoy, files smaller than 4k can be inlined.

Problems that Kubernetes solves that web servers no longer need to solve:
- Scaling horizontally

Problems that Ingress and Envoy solves that web servers no longer need to solve:
- Path rewrites
- TLS termination

New needs when static content goes "immutable infrastructure":
- Predictable memory footprint
- Prometheus metrics export

## How to use

See [example](./tests) [Dockerfile](./tests/html01/Dockerfile).

## Mime types

By extension, stdlib: https://pkg.go.dev/mime#TypeByExtension

Detection:
https://pkg.go.dev/github.com/gabriel-vasile/mimetype

## Client side caching

TODO 304 responses on If-None-Match

## Compression

TODO but we could obviously gzip assets at build time, preferrably per mime type.

### Compression optional by Accept header

We could compress at build time to save CPU at runtime.
Content would contain both uncompressed and compressed files, to select based on header matching. The complexity might not be warranted compared to Envoy [Compressor](https://www.envoyproxy.io/docs/envoy/v1.21.1/api-v3/extensions/filters/http/compressor/v3/compressor.proto.html?highlight=http%20filters)

## Adding dynamic content

Envoy was boarn a proxy: use a sidecar!

## Response transformation

A lightweight form of dynamic content is to [interpolate at response time](https://github.com/kris-nova/bjorno). TODO verify that direct_response works with for example [Lua](https://www.envoyproxy.io/docs/envoy/v1.21.1/api-v3/extensions/filters/http/lua/v3/lua.proto.html?highlight=lua) or even [Wasm](https://www.envoyproxy.io/docs/envoy/v1.21.1/configuration/listeners/network_filters/wasm_filter.html?highlight=wasm).

## Benchmarks

If performance matters consider using a CDN.

For our use cases we actually only care about the ease of operations,
but if anyone has benchmarks we're of course curious.

## Dev loop

```
DEBUG=true RUN_OPTS="--rm" ./test.sh

docker stop envoystatic-test
DEBUG=true ./test.sh
```

## References

https://github.com/envoyproxy/envoy/issues/378

- https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto.html?highlight=route_config
  - `rds` https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto.html?highlight=route_config#envoy-v3-api-msg-extensions-filters-network-http-connection-manager-v3-rds
    - `config_source` `path: /etc/envoy...`
