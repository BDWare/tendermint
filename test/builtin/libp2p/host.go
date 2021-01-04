// Copyright (c) 2020 The BDWare Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can be
// found in the LICENSE file.

package libp2p

import (
	"context"
	"github.com/libp2p/go-libp2p"
	"io/ioutil"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-tcp-transport"

	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/p2p"
)

func NewP2PHost(ctx context.Context, cfg *cfg.Config) (host.Host, error) {
	transports := libp2p.ChainOptions(
		libp2p.Transport(tcp.NewTCPTransport),
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

	var listenAddrs libp2p.Option
	tdmAddr, err := p2p.NewNetAddressString(cfg.P2P.ListenAddress)
	if err != nil {
		listenAddrs = libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/26656",
			//"/ip4/0.0.0.0/tcp/0/ws",
		)
	} else {
		listenAddrs = libp2p.ListenAddrs(p2p.MultiaddrFromNetAddress(*tdmAddr))
	}

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
		routing,
		id,
	)
	if err != nil {
		panic(err)
	}

	return host, nil
}
