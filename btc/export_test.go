package btc

import (
	"context"

	"github.com/ebiiim/btc-gateway/model"
)

func DryRun(b bool)                                                { dryRun = b }
func (b *BitcoinCLI) BinPath() string                              { return b.binPath }
func (b *BitcoinCLI) BTCNet() model.BTCNet                         { return b.btcNet }
func (b *BitcoinCLI) RPCAddr() string                              { return b.rpcAddr }
func (b *BitcoinCLI) RPCPort() string                              { return b.rpcPort }
func (b *BitcoinCLI) RPCUser() string                              { return b.rpcUser }
func (b *BitcoinCLI) RPCPassword() string                          { return b.rpcPassword }
func (b *BitcoinCLI) ConnArgs() []string                           { return b.connArgs() }
func (b *BitcoinCLI) Run(ctx context.Context, args []string) error { return b.run(ctx, args) }
