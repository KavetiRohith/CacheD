# CacheD

This is an example of using the Hashicorp Raft implementation, Raft is a distributed consensus protocol that ensures all nodes in a cluster agree on the state of an arbitrary state machine despite potential failures or network partitions. This type of consensus is crucial for building systems that can withstand faults. The example here is a simple in-memory key-value store that can be run on Linux, OSX, and Windows and is useful for understanding both the Raft consensus protocol and Hashicorp's implementation of it. The implementation includes commands for setting and reading keys through a TCP binding address, and instructions for building and running a cluster.

## Reading and writing keys

The implementation is a very simple in-memory key-value store. You can SET a key by sending the following message to the TCP bind address (which defaults to `localhost:3001`) using netcat, the provided client or any other TCP Client:

```bash
nc 127.0.0.1 3001
SET foo bar
Success
```

You can read the value for a key like this:

```bash
nc 127.0.0.1 3001
GET foo
bar
```

## Instructions to run a cluster

Building CacheD requires a recent version of Go.
Starting and running a CacheD cluster is easy. Download and build CacheD by running the following:

```bash
git clone https://github.com/KavetiRohith/CacheD.git
cd CacheD
make build
```

launch your first node by running the following:

```bash
./bin/CacheD -id node1 -tcpAddr localhost:3001 -rAddr localhost:4001 -raftDir ./node1
```

We can now SET a key and GET it's value

```bash
nc 127.0.0.1 3001
SET foo bar
Success
GET foo
bar
```

## Multi-node Clustering

The following is an example of launching a 3 node cluster.

## Walkthrough

You should start the first node like so:

```
./bin/CacheD -id node1 -tcpAddr localhost:3001 -rAddr localhost:4001 -raftDir ./node1
```

This node will start up and become leader of a single-node cluster.

Next, start the second node as follows:

```
./bin/CacheD -id node2 -tcpAddr localhost:3002 -rAddr localhost:4002 -raftDir ./node2 -join localhost:3001
```

Finally, start the third node as follows:

```
./bin/CacheD -id node3 -tcpAddr localhost:3003 -rAddr localhost:4003 -raftDir ./node3 -join localhost:3001
```

_Specifically using ports 3000 and 4000 is not required. You can use other ports as per the availability._

Note how each node listens on its own address, but joins to the address of the leader node. The second and third nodes will start, join the with leader at `localhost:3001`, and a 3-node cluster will be formed.

Once joined, the new nodes now know about the key:

```bash
nc 127.0.0.1 3002
GET foo
bar

```

```bash
nc 127.0.0.1 3003
GET foo
bar

```

### Tolerating failure

Kill the leader process and watch as one of the other nodes gets elected as the leader. The keys are still available for query on the other nodes, and we can set keys via the new leader. Furthermore, when the first node is restarted, it rejoins the cluster and learn about any updates that occurred while it was down.

A 3-node cluster can tolerate the failure of a single node, but a 5-node cluster can tolerate the failure of two nodes. But 5-node clusters require that the leader contact a larger number of nodes before any change e.g. setting a key's value, can be considered committed.

### Leader-forwarding

Automatically forwarding requests to set/delete keys to the current leader is not implemented. The client must always send requests to change a key to the leader or an error will be returned.
