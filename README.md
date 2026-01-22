# Eden Testnet EV-Node

This repository contains a fork of the [ev-node evm/single testapp](https://github.com/evstack/ev-node/tree/main/apps/evm/single) specifically for the Eden Testnet.

It adds the following features:

- Changed default chain-id to `edennet-2`
- Changed default app home directory to `~/.eden-testnet`
- Changed command name to `eden-testnet`
- Maintained backward compat for operators with deprecated flags (`--evnode.signer.passphrase` and `--evnode.evm.jwt-secret`)

## Considerations

While syncing from DA only, with the Eden testnet, you will experience a stale block sync at height **`2403026`**.
This is due to a manual operation performed on the sequencer to submit a missing blob. That blob has been submitted at DA height `8254985`.
Until the syncer has reached that height, **do not stop the node, nor clear the cache**.
