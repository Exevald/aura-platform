#!/bin/sh
set -u

shutdown() {
    if [ -f /tmp/registerer.pid ]; then
        kill "$(cat /tmp/registerer.pid)" 2>/dev/null || true
        wait "$(cat /tmp/registerer.pid)" 2>/dev/null || true
    fi
    exit 0
}

trap shutdown TERM INT

while true; do
    /registerer service &
    child=$!
    echo "$child" >/tmp/registerer.pid
    wait "$child" || true
    sleep 0.1
done
