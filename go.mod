module github.com/tendermint/tendermint

go 1.15

replace (
	github.com/libp2p/go-libp2p-kad-dht => github.com/bdware/go-libp2p-kad-dht v0.11.0-bdw
	github.com/libp2p/go-libp2p-kbucket => github.com/bdware/go-libp2p-kbucket v0.4.7-bdw
)

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/ChainSafe/go-schnorrkel v0.0.0-20200405005733-88cbf1b4c40d
	github.com/Workiva/go-datastructures v1.0.52
	github.com/bdware/go-datastore v0.4.5-alpha.2
	github.com/bdware/go-ds-leveldb v0.4.3-alpha.2
	github.com/bdware/tm-db-go-datastore v0.1.1
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/confio/ics23/go v0.6.3
	github.com/cosmos/iavl v0.15.0-rc5
	github.com/dgraph-io/badger v1.6.1
	github.com/fortytw2/leaktest v1.3.0
	github.com/go-kit/kit v0.10.0
	github.com/go-logfmt/logfmt v0.5.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/gorilla/websocket v1.4.2
	github.com/gtank/merlin v0.1.1
	github.com/hdevalence/ed25519consensus v0.0.0-20200813231810-1694d75e712a
	github.com/libp2p/go-buffer-pool v0.0.2
	github.com/libp2p/go-libp2p v0.12.0
	github.com/libp2p/go-libp2p-core v0.7.0
	github.com/libp2p/go-libp2p-kad-dht v0.11.1
	github.com/libp2p/go-libp2p-swarm v0.3.1
	github.com/libp2p/go-libp2p-yamux v0.4.1 // indirect
	github.com/libp2p/go-tcp-transport v0.2.1
	github.com/magiconair/properties v1.8.1
	github.com/minio/highwayhash v1.0.1
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.8.0
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0
	github.com/rs/cors v1.7.0
	github.com/sasha-s/go-deadlock v0.2.0
	github.com/snikch/goodman v0.0.0-20171125024755-10e37e294daa
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/tendermint/tm-db v0.6.3
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	golang.org/x/sys v0.0.0-20201231184435-2d18734c6014 // indirect
	google.golang.org/grpc v1.33.2
)
