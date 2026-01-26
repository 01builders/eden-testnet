# Eden Testnet EV-Node

## Considerations

While syncing from DA only, with the Eden testnet, you will experience a stale block sync at height **`2403026`**.
This is due to a manual operation performed on the sequencer to submit a missing blob. That blob has been submitted at DA height `8254985`.
Until the syncer has reached that height, **do not stop the node, nor clear the cache**.
