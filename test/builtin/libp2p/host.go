// Copyright (c) 2020 The BDWare Authors. All rights reserved.
// Use of this source code is governed by Apache License 2.0 that can be
// found in the LICENSE file.

package libp2p

import (
	"context"
	"io/ioutil"

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

	cfg "github.com/bdware/tendermint/config"
	"github.com/bdware/tendermint/p2p"
)

func NewP2PHost(ctx context.Context, cfg *cfg.Config) (host.Host, error) {
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

	var listenAddrs libp2p.Option
	tdmAddr, err := p2p.NewNetAddressString(cfg.P2P.ListenAddress)
	if err != nil {
		listenAddrs = libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/26656",
			//"/ip4/0.0.0.0/tcp/0/ws",
		)
	} else {
		listenAddrs = libp2p.ListenAddrs(tdmAddr.Multiaddr())
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
