# Eden Testnet EV-Node

This repository contains a fork of the [ev-node evm/single testapp](https://github.com/evstack/ev-node/tree/main/apps/evm/single) specifically for the Eden Testnet.

It adds the following features:

- Namespace migration support
  - Eden testnet had a migration from the default namespace `rollkit-headers` and `rollkit-data` to its own namespaces: `eden_testnet_header` and `eden_testnet_data`.
  - The migration height is backed directly in the binary to avoid having to restart a sync node.
- Changed default chain-id to `edennet-2`
- Changed default app home directory to `~/.eden-testnet`
- Changed command name to `eden-testnet`

It removed the following features:

- Deleted rollback command until https://github.com/celestiaorg/go-header/pull/347 is stable and merged.

## Considerations

While syncing from DA only, with the Eden testnet, you will experience a stale block sync at height **`2403026`**.
This is due to a manual operation performed on the sequencer to submit a missing blob. That blob has been submitted at DA height `8254985`.
Until the syncer has reached that height, **do not stop the node, nor clear the cache**.
