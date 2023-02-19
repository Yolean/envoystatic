#!/usr/bin/env bash
[ -z "$DEBUG" ] || set -x
set -eo pipefail

[ -z "$HOST" ] && echo "Test HOST is required" && exit 1

[ -n "CURL_OPTS" ] || [ -z "$DEBUG" ] || CURL_OPTS="-v"

curl -v $CURL_OPTS -f $HOST/index.html

curl -I $CURL_OPTS -f $HOST/index.html | grep 'text/html'

curl -I $CURL_OPTS -f $HOST/script.js | grep 'ETag'

curl -v $CURL_OPTS -f $HOST/script.js \
  -H 'If-None-Match: "63314764b32e0f86ebc1b32a734cba2dabc4945b7897fc024f37f0bf16ed4226"' \
