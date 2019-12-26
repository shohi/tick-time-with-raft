# Tick time with Raft

Every server runs the same time tick, but based on its ID, every server only shows some seconds marks in the log.

Every node contains all the information but the responsibility of show to the log (the execution) is distributed.

When a new server is added the responsibility is redistributed. (TODO: update)

**Start leader**

```bash
# node a
go run main.go -id a -raft-dir tmp/node_a -raft-addr :6001 -addr :9001

```

**Start peers**

```bash

# node b
go run main.go -id b -raft-dir tmp/node_b -raft-addr :6002 -addr :9002 -join :9001

# node c
go run main.go -id c -raft-dir tmp/node_c -raft-addr :6003 -addr :9003 -join :9001

```

## Reference

1. Go raft library, https://github.com/hashicorp/raft
2. Raft description, https://raft.github.io/
3. Using hashicorp/raft, https://github.com/otoolep/hraftd
