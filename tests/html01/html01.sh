#!/usr/bin/env bash
[ -z "$DEBUG" ] || set -x
set -eo pipefail

[ -z "$HOST" ] && echo "Test HOST is required" && exit 1

[ -n "CURL_OPTS" ] || [ -z "$DEBUG" ] || CURL_OPTS="-v"

curl $CURL_OPTS -f $HOST/index.html
