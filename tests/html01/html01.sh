#!/usr/bin/env bash
[ -z "$DEBUG" ] || set -x
set -eo pipefail

[ -z "$HOST" ] && echo "Test HOST is required" && exit 1

[ -n "CURL_OPTS" ] || [ -z "$DEBUG" ] || CURL_OPTS="-v"

curl -v $CURL_OPTS -f $HOST/index.html

curl -I $CURL_OPTS -f $HOST/index.html | grep 'text/html'
