#!/bin/sh

set -ex

mongod --storageEngine=inMemory \
    --inMemorySizeGB=1 \
    --bind_ip=127.0.0.1 \
    --journal \
    --fork \
    --logpath=/var/log/mongod.log
server
