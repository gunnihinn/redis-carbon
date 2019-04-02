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

trap cleanup EXIT

function send {
    echo "foo $1 -1" | netcat -c 127.0.0.1 2003
}

function test_send {
    send 1
    send 2
    send 3

    len=$(redis-cli XLEN metric:foo)

    if [[ "$len" -ne 3 ]]; then
        echo "Expected 3 elements in stream, got $len:"
        redis-cli XRANGE metric:foo - +
        exit 1
    fi
}

function test_load {
    total=1000
    for i in $(seq 1 "$total"); do
        send "$i" &
    done

    exp=$((total + 3))

    sleep 1

    sent=$(curl -s http://127.0.0.1:8080/debug/vars | jq '.point_total')
    if [[ "$sent" -ne "$exp" ]]; then
        echo "Only sent $sent / $exp points"
        exit 2
    fi

    fail=$(curl -s http://127.0.0.1:8080/debug/vars | jq '.point_errors')
    if [[ "$fail" -ne 0 ]]; then
        echo "Failed to send $fail out of $exp points"
        exit 2
    fi

    recv=$(redis-cli XLEN metric:foo)
    if [[ "$recv" -ne "$sent" ]]; then
        echo "Redis only received $rect / $sent points"
        exit 2
    fi
}

setup

echo "Test 1"
test_send

echo "Test 2"
test_load
