# Solana Cluster Manager

Tooling to manage private clusters of Solana nodes.

## Architecture

### Snapshot management

**[Twitter 🧵](https://twitter.com/terorie_dev)**

Snapshot management tooling enables efficient peer-to-peer transfers of accounts database archives.

![Snapshot Fetch](./docs/snapshots.png)

**Scraping** (Flow A)

Snapshot metadata collection runs periodically similarly to Prometheus scraping.

Each cluster-aware node runs a lightweight `solana-cluster sidecar` agent providing telemetry about its snapshots.

The `solana-cluster tracker` then connects to all sidecars to assemble a complete list of snapshot metadata.
The tracker is stateless so it can be replicated.
Service discovery is available through HTTP and JSON files. Consul SD support is planned.

Side note: Snapshot sources are configurable in stock Solana software but only via static lists.
This does not scale well with large fleets because each cluster change requires updating the lists of all nodes.

**Downloading** (Flow B)

When a Solana node needs to fetch a snapshot remotely, the tracker helps it find the best snapshot source.
Nodes will download snapshots directly from the sidecars of other nodes.

### TPU & TVU

Not yet public. 🚜 Subscribe to releases! ✨

## Motivation

Blockdaemon manages one of the largest Solana validator and RPC infrastructure deployments to date, backed by a custom peer-to-peer backbone.
This repository shares our tools for performance and sustainability improvements.

When Solana validators first start, they have to retrieve and validate hundreds of gigabytes of state data from a remote node.
During normal operation, validators stream at least 500 Mbps of traffic in either direction.

For Solana infra operators that manage more than node (not to mention hundreds), this cost currently scales linearly as well.
Unmodified Solana deployments treat their cluster peers the same as any other.
This can end in a large number of streams between globally dispersed validators.

This is obviously inefficient. 10 Gbps connectivity is cheap and abundant locally within data centers.
In contrast, major public clouds (who shall not be named) charge egregious premiums on Internet traffic.

The solution: Co-located Solana validators that are controlled by the same entity should also behave as one entity.

Leveraging internal connectivity to distribute blockchain data can
reduce public network _leeching_ and increase total cluster bandwidth.

Authenticated internal connectivity allows delegation of expensive calculations and re-use of results thereof.
Concretely, the amount write-heavy snapshot creation & verification procedures per node can decrease as the cluster scales out.
