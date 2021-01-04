#!/bin/sh

# Be careful, this will overwrite existing config files.
TMHOME="$PWD/example" tendermint init

cp example/data/priv_validator_state.json example/data/priv_validator_state.json.bak