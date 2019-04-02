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

trap EXIT cleanup
trap ERROR cleanup


