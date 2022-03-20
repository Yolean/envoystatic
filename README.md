# envoystatic

Rethinking the static web server for immutable infrastructure.

Container images  has a build step. The old static file server requirements no longer apply.
We can instead pre-process the directory structure to be hosted.

While other static file servers mount a volume,
here we build a container that represents the content at a specific state.
There is no longer a need to check the underlying file system for changes.

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

The [yolean/envoystatic](yolean/envoystatic)`:tooling-[gitref]`
image is for the build step.

The [yolean/envoystatic](yolean/envoystatic)`:envoy-[gitref]`
is slightly more opinionated than the official envoy image.
- Sets loglevel and xDS names as default command.
- Uses subfolders to `/etc/envoy` for bootstrap and xDS
  to allow separate mounts if desired.
- Runs as envoy's nonroot by default
- TODO distroless once envoy's distroless becomes multi-arch.

## Disclaimer

This project is a Proof-of-concept and not necessarily suitable for production.
It explores if build time preparation of an HTTP server is preferrable to serving a directory as-is.
The most notable caveat is that
[Envoy's warning on direct responses](https://www.envoyproxy.io/docs/envoy/v1.21.1/api-v3/config/route/v3/route.proto.html?highlight=max_direct_response_body_size_bytes)
applies because we [increase the limit](https://github.com/envoyproxy/envoy/pull/14778) to an arbitrarily high value.

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
# Start the first test container for exploration
DEBUG=true RUN_OPTS="--rm" ./test.sh

# Run all tests
docker stop envoystatic-test
DEBUG=true NOPUSH=true ./test.sh

# For iterating with a local downstream docker build
DEBUG=true NOPUSH=true PLATFORM="--load" ./test.sh
```

## References

https://github.com/envoyproxy/envoy/issues/378

- https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto.html?highlight=route_config
  - `rds` https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto.html?highlight=route_config#envoy-v3-api-msg-extensions-filters-network-http-connection-manager-v3-rds
    - `config_source` `path: /etc/envoy...`

## Misc TODOs

- Validate against RDS errors at runtime, at least in test. Ideally make them fatal.
- Adapt `max_direct_response_body_size_bytes` to the size of the largest processed file.
- Update bootstrap config according to deprecation warnings.
- A `skaffold dev` loop with output dir and route.yaml sync, for downstream work.
  - For example with `npm run build` in a Nextjs project
- Favicon support
- Redirect to index.html per subdir
