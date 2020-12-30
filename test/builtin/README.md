# Reference
[Creating a built-in application in Go](https://docs.tendermint.com/master/tutorials/go-built-in.html)

# Build

```shell
go build
```

# Generate Config Files

Only need do this once.

Use `tendermint init`:

1. Run `genconf.sh` in each node.

   Note this will overwrite existing config files.

2. Collect all peer node ids(you can use `node_id.sh`), hosts and ports and fill them in p2p/persistent_peers field of all nodes' config.toml.

Or use `tendermint testnode`:

1. Run `tendermint testnode` in any node.

2. Distribute generated test config files to each node manually.

# Run

```shell
./run.sh
```

# Send Tx & Query

See [Getting Up and Running](https://docs.tendermint.com/master/tutorials/go-built-in.html#_1-5-getting-up-and-running).

# Clean

```shell
./clean
```

This will remove all app data and reset peer state.

# Badger Storage

default location: `/tmp/tendermint/test-builtin/badger`
