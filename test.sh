#!/bin/bash

set euo pipefail

setup() {
    make clean all
    redis-carbon &
    sudo docker run --name test-redis -p 127.0.0.1:6379:6379 -d redis
}

cleanup() {
    pkill -f redis-carbon
    sudo docker stop test-redis
}

send() {
    echo "foo $1 -1" | netcat -c 127.0.0.1 6379
}

trap EXIT cleanup
trap ERROR cleanup

setup

send 1
send 2
send 3

redis-cli 'XRANGE metric:foo - +'
