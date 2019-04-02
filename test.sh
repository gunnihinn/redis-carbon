#!/bin/bash

set -euo pipefail

function setup {
    make all > /dev/null
    ./redis-carbon &
    sudo docker run --name test-redis -p 127.0.0.1:6379:6379 -d redis > /dev/null
}

function cleanup {
    pkill -f redis-carbon
    sudo docker stop test-redis > /dev/null
    sudo docker rm test-redis > /dev/null
}

function send {
    echo "foo $1 -1" | netcat -c 127.0.0.1 2003
}

trap cleanup EXIT

setup

send 1
send 2
send 3

len=$(redis-cli XLEN metric:foo)

if [[ "$len" -ne 3 ]]; then
    echo "Expected 3 elements in stream, got $len:"
    redis-cli XRANGE metric:foo - +
    exit 1
fi
