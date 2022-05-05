# Solana Cluster Manager

Tooling to manage private clusters of Solana nodes.

<a href="https://pkg.go.dev/go.blockdaemon.com/solana/cluster-manager"><img src="https://pkg.go.dev/badge/go.blockdaemon.com/solana/cluster-manager.svg" alt="Go Reference"></a>
<img src="https://github.com/Blockdaemon/solana-cluster/actions/workflows/build.yml/badge.svg?branch=main">
<img src="https://github.com/Blockdaemon/solana-cluster/actions/workflows/test.yml/badge.svg?branch=main">

[Issue Tracker](https://github.com/orgs/Blockdaemon/projects/1/views/1)


## Deployment

### Building from source

Requirements: Go 1.18 & build essentials.

```shell
go mod download
go build -o ./solana-cluster .
```

### Docker Image

Find amd64 and arm64 Docker images for all releases on [GitHub Container Registry](https://github.com/Blockdaemon/solana-cluster/pkgs/container/solana-cluster-manager).

```shell
docker pull ghcr.io/blockdaemon/solana-cluster-manager
docker run \
  --network-mode=host \
  -v /solana/ledger:/ledger:ro \
  ghcr.io/blockdaemon/solana-cluster-manager \
  sidecar --ledger /ledger
```

## Usage

```
$ solana-cluster sidecar --help

Runs on a Solana node and serves available snapshot archives.
Do not expose this API publicly.

Usage:
  solana-snapshots sidecar [flags]

Flags:
      --interface string    Only accept connections from this interface
      --ledger string       Path to ledger dir
      --port uint16         Listen port (default 13080)
```

```
$ solana-cluster tracker --help

Connects to sidecars on nodes and scrapes the available snapshot versions.
Provides an API allowing fetch jobs to find the latest snapshots.
Do not expose this API publicly.

Usage:
  solana-snapshots tracker [flags]

Flags:
      --config string            Path to config file
      --internal-listen string   Internal listen URL (default ":8457")
      --listen string            Listen URL (default ":8458")
```

```
$ solana-cluster fetch --help

Fetches a snapshot from another node using the tracker API.

Usage:
  solana-snapshots fetch [flags]

Flags:
      --download-timeout duration   Max time to try downloading in total (default 10m0s)
      --ledger string               Path to ledger dir
      --max-slots uint              Refuse to download <n> slots older than the newest (default 10000)
      --min-slots uint              Download only snapshots <n> slots newer than local (default 500)
      --request-timeout duration    Max time to wait for headers (excluding download) (default 3s)
      --tracker string              Download as instructed by given tracker URL
```

```
$ solana-cluster mirror --help

Periodically mirrors snapshots from nodes to an S3-compatible data store.
Specify credentials via env $AWS_ACCESS_KEY_ID and $AWS_SECRET_ACCESS_KEY

Usage:
  solana-snapshots mirror [flags]

Flags:
      --refresh duration   Refresh interval to discover new snapshots (default 30s)
      --s3-bucket string   Bucket name
      --s3-prefix string   Prefix for S3 object names (optional)
      --s3-region string   S3 region (optional)
      --s3-secure          Use secure S3 transport (default true)
      --s3-url string      URL to S3 API
      --tracker string     URL to tracker API
```

## Architecture

### Snapshot management

**[Twitter 🧵](https://twitter.com/terorie_dev/status/1520289936611725312)**

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
This repository shares our performance and sustainability optimizations.

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
