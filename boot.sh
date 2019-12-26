#!/usr/bin/env bash
set -x

# steps to test raft cluster locally
# 1. prepare test folders
# make prepare
#
# 1. start raft node
#
# node a: (*seed*)
#    go run main.go -id a -raft-dir tmp/node_a -raft-addr :6001 -addr :9001
#
# node b:
#    go run main.go -id b -raft-dir tmp/node_b -raft-addr :6002 -addr :9002 -join :9001
#
# node c:
#    go run main.go -id c -raft-dir tmp/node_c -raft-addr :6003 -addr :9003 -join :9001
