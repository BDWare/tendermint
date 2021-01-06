// Copyright (c) 2020 The BDWare Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can be
// found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/dgraph-io/badger"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/test/builtin/libp2p"
)

var home = os.Getenv("HOME")
var configFile string
var badgerPath string

func init() {
	flag.StringVar(&configFile, "config", home+"/.tendermint/config/config.toml", "Path to config.toml")
	flag.StringVar(&badgerPath, "badger", "/tmp/tendermint/test-builtin/badger", "Path to badger db")
}

func main() {
	if _, err := os.Stat(badgerPath); os.IsNotExist(err) {
		err = os.MkdirAll(badgerPath, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create badger db dir of %v: %v", badgerPath, err)
			os.Exit(1)
		}
	}
	db, err := badger.Open(badger.DefaultOptions(badgerPath))
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
	if logger, err = tmflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel()); err != nil {
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	// read private validator
	pv := privval.LoadFilePV(
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
	)

	var nodeKey *p2p.NodeKey
	if !config.P2P.Libp2p {
		// read node key
		if nodeKey, err = p2p.LoadNodeKey(config.NodeKeyFile()); err != nil {
			return nil, fmt.Errorf("failed to load node's key: %w", err)
		}
	}

	// create libp2p host
	var host host.Host
	if config.P2P.Libp2p {
		if host, err = libp2p.NewP2PHost(context.Background(), config); err != nil {
			return nil, fmt.Errorf("failed to create new libp2p host: %w", err)
		}
		logger.Info("Libp2p host created", "host.ID", host.ID(), "host.Addrs:", host.Addrs())
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
