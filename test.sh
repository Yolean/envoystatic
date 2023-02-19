#!/usr/bin/env bash
[ -z "$DEBUG" ] || set -x
set -eo pipefail

[ -n "$BUILDX" ] || BUILDX=buildx
[ "$($BUILDX version)" != "github.com/docker/buildx v0.10.0 876462897612d36679153c3414f7689626251501" ] && echo "Unexpected buildx version" && exit 1
# BUILDX="$BUILDX --progress=plain"

echo "==> Collecting build settings"
[ -n "$PLATFORMS" ] || PLATFORMS="linux/amd64,linux/arm64/v8"
[ -n "$PLATFORM" ] || PLATFORM="--platform=$PLATFORMS"
[ -z "$REGISTRY" ] || PREFIX="$REGISTRY/"
[ -n "$PUSH" ] || [ "$NOPUSH" = "true" ] || PUSH="--push"

SOURCE_COMMIT=$(git rev-parse --verify --short HEAD 2>/dev/null || echo '')
if [[ ! -z "$SOURCE_COMMIT" ]]; then
  GIT_STATUS="$(git status --untracked-files=normal --porcelain=v2)"
  if [[ ! -z "$GIT_STATUS" ]]; then
    echo "# Tagging as -dirty due to:"
    echo "$GIT_STATUS"
    SOURCE_COMMIT="$SOURCE_COMMIT-dirty"
    [ PUSH != "--push" ] || PUSH=""
  fi
fi
# BUILD_ARGS="$BUILD_ARGS --build-arg SOURCE_COMMIT=$SOURCE_COMMIT"

# [ -n "$ENVOY_VERSION" ] || ENVOY_VERSION="v1.25.1"
# BUILD_ARGS="$BUILD_ARGS --build-arg ENVOY_VERSION=$ENVOY_VERSION"

echo "# Platform: $PLATFORM"
echo "# Image name prefix (i.e. registry, empty = default): $PREFIX"
echo "# Build args: $BUILD_ARGS"

echo "==> Building base image for tests, single platform"

for STAGE in tooling envoy; do
  IMAGE=yolean/envoystatic:$STAGE
  echo "==> $IMAGE"
  $BUILDX build $BUILD_ARGS -t $IMAGE .
done
unset IMAGE

exit 0

echo "==> Collecting test settings"
PORT=8080
HOST=http://localhost:$PORT
NAME=envoystatic-test
#HEALTH_CHECK_PATH=/
HEALTH_CHECK_PATH=/index.html
[ -n "$RUN_OPTS" ] || RUN_OPTS="-d --rm"

echo "==> Running unit tests"
go test ./...

echo "==> Running e2e tests"
[ -n "$TESTS" ] || TESTS='
html01
'
trap "echo Terminating; docker stop $NAME; exit" SIGINT

for TEST in $TESTS; do
  TEST=html01
  echo "==> Running test: $TEST"

  echo "==> Running local transform"
  tmpout=$(mktemp -d)
  go run ./cmd/envoystatic route --in=./tests/$TEST --out="$tmpout/docroot" --rdsyaml=-
  ls -1 $tmpout/docroot

  echo "==> Building downstream image"
  TESTIMAGE=envoystatic-test-$TEST
  # hack to run dependent builds without base image push
  cat Dockerfile ./tests/$TEST/Dockerfile \
    | sed 's|yolean/envoystatic:||' \
    | sed 's|--from=0|--from=3|' \
    | sed 's|COPY \. |COPY --from=testsource . |' \
    | tee Dockerfile.test \
    | $BUILDX build $BUILD_ARGS -t $TESTIMAGE --load --build-context testsource=tests/$TEST -f - .

  echo "==> Printing generated route config for $TEST"
  docker run --rm --entrypoint cat $TESTIMAGE /etc/envoy/rds/route.yaml

  echo "==> Starting webserver $TEST using container name $NAME"
  docker run $RUN_OPTS -p $PORT:8080 --name $NAME $TESTIMAGE

  echo "==> Checking readiness using propbe-ish curl to $HOST$HEALTH_CHECK_PATH"
  until sleep 1 && curl -f -H 'User-Agent: kube-probe/mock' \
    -o /dev/null -w "%{http_code}\n" $HOST$HEALTH_CHECK_PATH; do
      echo "Not ready, current last log line:"
      docker logs --tail=1 $NAME
  done

  echo "==> Checking for warnings or errors in envoy logs"
  docker logs $NAME 2>&1 | grep -E '(warning|error|critical)' || echo "# None"
  echo "==> Startig test $TEST"
  HOST=$HOST ./tests/$TEST/$TEST.sh
  echo "==> Test exited ok, killing container"
  docker stop $NAME
done

echo "==> Tests passed, build and push $PLATFORM"

# docker buildx build --push --platform=$PLATFORMS
# docker buildx build $BUILDX_PUSH --progress=plain $PLATFORM \
#   -t yolean/docker-base -t ${PREFIX}yolean/envoystatic:$SOURCE_COMMIT$XTAG -

for STAGE in tooling envoy; do
  TAG="$STAGE-"
  # [ "$TAG" != "envoy-" ] || TAG="$TAG$ENVOY_VERSION-"
  [ "$TAG" != "-" ] || TAG=""
  IMAGE=${PREFIX}yolean/envoystatic:$TAG$SOURCE_COMMIT$XTAG
  LATEST=${PREFIX}yolean/envoystatic:$STAGE
  echo "==> $IMAGE"
  $BUILDX build $PUSH $PLATFORM $BUILD_ARGS --target $STAGE -t $IMAGE -t $LATEST .
done
unset IMAGE
