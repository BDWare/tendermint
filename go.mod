module github.com/bdware/tendermint

go 1.14

replace (
    github.com/libp2p/go-libp2p-kad-dht => github.com/bdware/go-libp2p-kad-dht v0.8.2-bdw.2
    github.com/libp2p/go-libp2p-kbucket => github.com/bdware/go-libp2p-kbucket v0.4.2-bdw.1
)

require (
	github.com/ChainSafe/go-schnorrkel v0.0.0-20200405005733-88cbf1b4c40d
	github.com/Workiva/go-datastructures v1.0.52
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/dgraph-io/badger v1.6.1
	github.com/fortytw2/leaktest v1.3.0
	github.com/go-kit/kit v0.10.0
	github.com/go-logfmt/logfmt v0.5.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.0
	github.com/gorilla/websocket v1.4.2
	github.com/gtank/merlin v0.1.1
	github.com/libp2p/go-buffer-pool v0.0.2
	github.com/libp2p/go-libp2p v0.9.6
	github.com/libp2p/go-libp2p-core v0.5.7
	github.com/libp2p/go-libp2p-kad-dht v0.8.2
	github.com/libp2p/go-libp2p-mplex v0.2.3
	github.com/libp2p/go-libp2p-secio v0.2.2
	github.com/libp2p/go-libp2p-swarm v0.2.6
	github.com/libp2p/go-libp2p-tls v0.1.3
	github.com/libp2p/go-libp2p-yamux v0.2.8
	github.com/libp2p/go-tcp-transport v0.2.0
	github.com/magiconair/properties v1.8.1
	github.com/minio/highwayhash v1.0.0
	github.com/multiformats/go-multiaddr v0.2.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0
	github.com/rs/cors v1.7.0
	github.com/snikch/goodman v0.0.0-20171125024755-10e37e294daa
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.6.3
	github.com/stretchr/testify v1.6.1
	github.com/tendermint/go-amino v0.14.1
	github.com/tendermint/tm-db v0.5.1
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20200519113804-d87ec0cfa476
	google.golang.org/grpc v1.28.1
)
