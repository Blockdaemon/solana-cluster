# Solana Snapshot Tools

This repo will contain scripts that enable efficient Solana snapshot creation and retrieval.

**Work In Progress!** – please come back later.

Brought to you by [Blockdaemon](https://blockdaemon.com/) and [Triton One](https://triton.one/#/).

## Goals

- Speed up RPC and validator node startup time
- Reduce NVMe wear by minimizing writes induced by snapshot creation

## Motivation

Solana nodes require snapshots of the accounts database to start up, which are expensive to create and go stale quickly.

Worse, NVMe SSDs get permanently damaged slightly whenever a snapshot gets created. This quickly escalates into multiple petabytes written and will permanently break a drive in a little over a year.

Finally, when no locally stored snapshot is found, a fallback mechanism kicks in to attempt to download a snapshot from a random peer on the public P2P network.

When operating multiple Solana nodes in the same data center, this creates obvious ineffiencies:
- Downloading snapshots from untrusted peers creates attack surface.
- Nodes should prefer pulling snapshots from peers in the same vicinity (DC) than selecting a random node.
- Many nodes will be generating equal snapshots at the same time when only one is required.

If we add cluster awareness to snapshot creation/retrieval, we can make the startup process more reliable, and node operation more sustainable in the long run.

## Measuring startup time

The startup phase is the time frame between the instruction for a node to start up, and when that node is ready to serve requests/validate. It is around 5 to 30 minutes in the best case.

In order to work towards minimizing the startup time, we first need to be able to measure it.

We know when a Solana node is started by the systemd unit activation time.

```shell
systemctl show solana --property=ActiveEnterTimestamp
```

It’s a bit more tricky to measure when startup finishes, especially for validators.
The `solana-validator` provides a hack:

The `--init-complete-file` flag makes the node touch a file when init completes.

The startup duration is therefore the modification time of the “complete file” minus the unit ActiveEnterTimestamp

## Snapshot creation

Solana nodes create snapshots on a fixed schedule (citation needed).

We will need to introduce an option into `solana-validator` to prevent automatic generation of snapshots and insert our own mechanism.

We add a new `systemd` timer that periodically stops a special "snapshot" node to create a snapshot on a preferred schedule.

```shell
solana-ledger-tool create-snapshot
```

For example one could achieve a "round-robin" snapshot schedule where nodes take turn generating snapshots.

## Snapshot management

The `solana-snapshot-service` program will then keep track of all created snapshots,
allowing us to always have a pointer to the latest available snapshot in our "cluster" of nodes.

This service is deliberately designed to be simple;
It serves a small HTTP API that simply redirects to the latest observed snapshot (307 Temporary Redirect).
This makes it safe to run redundantly for high availability.

Internally, this service runs a clock to perform the following steps every ~2 seconds:
- Discover all snapshot nodes via a service discovery backend.
- Send a HEAD HTTP request to all snapshot nodes to retrieve the modification dates of each snapshot.
- Choose the latest snapshot to be served for the redirect.

The service discovery mechanism is implementation defined –
the snapshot service will simply read in a text list of URLs.

## Snapshot retrieval

Retrieving a snapshot from a Solana node is trivial.

Consider the following pseudocode for `solana-snapshot-retrieve`.

```python
SNAPSHOT_MAX_AGE = 60 # seconds


def needs_snapshot_retrieval():
    # Check if we have a local snapshot.
    snap_info: Path = find_local_snapshot()
    if not snap_info.exists():
        return True
    
    # Check how old the local snapshot is.
    current_time = time.time()
    snap_time = snap_info.stat().st_mtime
    if current_time - snap_time > SNAPSHOT_MAX_AGE:
        return True
    
    # Everything checks out.
    return False


def download_snapshot():
    # Invoke cURL, retry three times.


if needs_snapshot_retrieval():
    try:
        download_snapshot()
    except e:
        print("Too bad")

sys.exit(0)
```

In English:
- Check if a snapshot exists, otherwise ignore
- Check if the snapshot is recent enough, otherwise ignore
- Invoke a download against the `solana-snapshot-service`
- The snapshot service will redirect us to the best node
- Directly transfer a snapshot from a snapshot node to a target node

In effect, when invoked, this tool ensures that a node has a recent snapshot locally available,
only downloading things if required.

The tool will be invoked whenever a Solana node starts through `ExecStartPre=` in the `solana.service` systemd unit.

To cover the worst-case scenario of having no snapshot available or an unrecoverable failure during snapshot retrieval,
we ignore any errors to be able to fall back to Solana node's gossip snapshot retrieval mechanism.
