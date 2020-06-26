[Creating a built-in application in Go](https://docs.tendermint.com/master/guides/go-built-in.html)

# build

```bash
go build
```

# generate config

use tendermint core

```bash
TMHOME="$PWD/example" tendermint init
```

or

```bash
tendermint testnode
```

# run

```bash
./tendermint_app -config="example/config/config.toml"
```

# badger storage

default:/tmp/badger
