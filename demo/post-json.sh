#!/bin/bash
curl -X POST \
  http://127.0.0.1:8181/v1/data/swisscom/example/allow \
  -d '{"input":{"method":"GET","path":["api","v1","employees","alice"],"user":"alice"}}' \
  -H 'Content-Type: application/json' | jq .
