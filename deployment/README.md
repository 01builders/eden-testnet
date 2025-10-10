# Eden Testnet Full Node Deployment

This guide describes how to run a full node for the Eden Testnet. There are two methods available: syncing from genesis or restoring from a database snapshot.

## Prerequisites

- Docker and Docker Compose installed
- Access to a Celestia light node RPC

## Method 1: Sync from Genesis

This method syncs the node from the beginning of the blockchain.

### Steps:

#### a) Configure Celestia Connection

Edit the `.env` file and set `DA_ADDRESS` to the address of your Celestia light node RPC:

```bash
DA_ADDRESS=http://your-celestia-node:26658
```

#### b) Start the Docker Compose Stack

Run the Docker Compose stack:

```bash
docker compose up -d
```

#### c) Wait for Initial Sync

Monitor the logs and wait for `ev-node` to sync to block **2217057**:

```bash
docker compose logs -f ev-node
```

#### d) Update Namespace Configuration

Once synced to block 2217057, stop the services:

```bash
docker compose down
```

Modify the `.env` file with the updated namespace values:

```bash
DA_HEADER_NAMESPACE=eden_testnet_header
DA_DATA_NAMESPACE=eden_testnet_data
```

#### e) Resume Synchronization

Start the `ev-node` again and wait for synchronization to the latest block:

```bash
docker compose up -d
docker compose logs -f ev-node
```

---

## Method 2: Restoration from snapshots

This method restores the node from databases snapshots for faster synchronization.

### Steps:

#### a) Initialize Nodes

Initialize the nodes with the Docker Compose stack:

```bash
docker compose up -d
```

#### b) Shut Down Nodes

Stop all running services:

```bash
docker compose down
```

#### c) Download and Extract Database Archives

Download the database archives and extract them to replace the `ev-node` and `ev-reth` databases.

- [ev-node database snapshot](https://prd-snapshots.fsn1.your-objectstorage.com/testnet/ev-node-fullnode-data.251009.tar.xz)
- [ev-reth database snapshot](https://prd-snapshots.fsn1.your-objectstorage.com/testnet/ev-reth-fullnode-data.251009.tar.xz)

**Note:** Ensure the extracted data replaces the existing database directories for both `ev-node` and `ev-reth`.

#### d) Update Namespace Configuration

Modify the `.env` file with the correct namespace values:

```bash
DA_HEADER_NAMESPACE=eden_testnet_header
DA_DATA_NAMESPACE=eden_testnet_data
```

#### e) Start and Synchronize

Start the `ev-node` again and wait for synchronization to the latest block:

```bash
docker compose up -d
docker compose logs -f ev-node
```

---

## Monitoring

To monitor your node's progress:

```bash
# View all logs
docker compose logs -f

# View specific service logs
docker compose logs -f ev-node
docker compose logs -f ev-reth
```
