package main

import (
	"github.com/bdware/tendermint/p2p"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	mplex "github.com/libp2p/go-libp2p-mplex"
	secio "github.com/libp2p/go-libp2p-secio"
	tls "github.com/libp2p/go-libp2p-tls"
	yamux "github.com/libp2p/go-libp2p-yamux"
	"github.com/libp2p/go-tcp-transport"
	"io/ioutil"

	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/dgraph-io/badger"
	"github.com/spf13/viper"

	abci "github.com/bdware/tendermint/abci/types"
	cfg "github.com/bdware/tendermint/config"
	tmflags "github.com/bdware/tendermint/libs/cli/flags"
	"github.com/bdware/tendermint/libs/log"
	nm "github.com/bdware/tendermint/node"
	"github.com/bdware/tendermint/privval"
	"github.com/bdware/tendermint/proxy"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "$HOME/.tendermint/config/config.toml", "Path to config.toml")
}

func main() {
	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open badger db: %v", err)
		os.Exit(1)
	}
	defer db.Close()
	app := NewKVStoreApplication(db)

	flag.Parse()

	node, err := newTendermint(app, configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(2)
	}

	node.Start()
	defer func() {
		node.Stop()
		node.Wait()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	os.Exit(0)
}

func newTendermint(app abci.Application, configFile string) (*nm.Node, error) {
	// read config
	config := cfg.DefaultConfig()
	config.SetRoot(filepath.Dir(filepath.Dir(configFile)))
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("viper failed to read config file: %w", err)
	}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("viper failed to unmarshal config: %w", err)
	}
	if err := config.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("config is invalid: %w", err)
	}

	// create logger
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	var err error
	logger, err = tmflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel())
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	// read private validator
	pv := privval.LoadFilePV(
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
	)

	var nodeKey *p2p.NodeKey
	if !config.P2P.Libp2p {
		//read node key
		nodeKey, err = p2p.LoadNodeKey(config.NodeKeyFile())
		if err != nil {
			return nil, fmt.Errorf("failed to load node's key: %w", err)
		}
	}

	var host host.Host
	// create libp2p host
	if config.P2P.Libp2p {
		host, err = newP2PHost(context.Background(), config)
		if err != nil {
			return nil, fmt.Errorf("failed to create new libp2p host: %w", err)
		}
		fmt.Println("host.ID:", host.ID())
		fmt.Println("host.Addrs:", host.Addrs())
	}

	// create node
	node, err := nm.NewNode(
		config,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		nm.DefaultGenesisDocProviderFunc(config),
		nm.DefaultDBProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger,
		host)
	if err != nil {
		return nil, fmt.Errorf("failed to create new Tendermint node: %w", err)
	}

	return node, nil
}

func newP2PHost(ctx context.Context, cfg *cfg.Config) (host.Host, error) {
	transports := libp2p.ChainOptions(
		libp2p.Transport(tcp.NewTCPTransport),
	)

	muxers := libp2p.ChainOptions(
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
		libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
	)

	keyBytes, err := ioutil.ReadFile(cfg.NodeKeyFile())
	if err != nil {
		return nil, err
	}

	pk, err := crypto.UnmarshalPrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}
	id := libp2p.Identity(pk)

	security := libp2p.ChainOptions(libp2p.Security(secio.ID, secio.New),
		libp2p.Security(tls.ID, tls.New))

	// TODO: modify this after refactor config address type
	listenAddrs := libp2p.ListenAddrStrings(
		"/ip4/0.0.0.0/tcp/26656",
		//"/ip4/0.0.0.0/tcp/0/ws",
	)

	var dht *kaddht.IpfsDHT
	newDHT := func(h host.Host) (routing.PeerRouting, error) {
		var err error
		dht, err = kaddht.New(ctx, h)
		//dht, err = dual.New(ctx, h, kaddht.ProtocolPrefix("/tdmApp"), kaddht.Mode(kaddht.ModeServer))
		return dht, err
	}
	routing := libp2p.Routing(newDHT)

	host, err := libp2p.New(
		ctx,
		transports,
		listenAddrs,
		muxers,
		security,
		routing,
		id,
	)
	if err != nil {
		panic(err)
	}

	return host, nil
}
