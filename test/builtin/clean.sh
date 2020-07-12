#!/bin/bash

# badger db
rm -rf /tmp/tendermint/test-builtin/badger > /dev/null 2>&1

# addrbook
rm example/config/addrbook.json  > /dev/null 2>&1

# db
rm -rf example/data/*.db > /dev/null 2>&1

# wal
rm -rf example/data/*.wal > /dev/null 2>&1

# reset priv validator state using backup file
cp example/data/priv_validator_state.json.bak example/data/priv_validator_state.json
