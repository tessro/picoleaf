#!/usr/bin/env bash

set -e

HOST=$( (timeout 2 stdbuf -o0 dns-sd -Z _nanoleafapi | grep -o '\d* \w*\-.*\.local') | awk '{ printf "%s:%s", $2, $1 }')

AUTH=$(curl -sLX POST http://"${HOST}"/api/v1/new | grep -o ':"[^"]*' )
AUTH="${AUTH:2}"

printf 'host=%s' "${HOST}"
echo
printf 'access_token=%s' "${AUTH}"
echo

