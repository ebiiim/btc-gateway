package btc

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ebiiim/btc-gateway/model"
)

type cliCmd string

const (
	cmdPing = "ping"
	// cmdGetNewAddress                = "getnewaddress"
	// cmdListUnspent                  = "listunspent"
	// cmdCreateRawTransaction         = "createrawtransaction"
	// cmdSignRawTransactionWithWallet = "signrawtransactionwithwallet"
	// cmdSendRawTransaction           = "sendrawtransaction"
)

type cliArg string

const (
	argChain       = "-chain="
	argRPCAddr     = "-rpcconnect="
	argRPCPort     = "-rpcport="
	argRPCUser     = "-rpcuser="
	argRPCPassword = "-rpcpassword="
)

type exitCode int

const (
	exitOK  = 0
	exitERR = 1
)

// Errors
var (
	errUnexpectedExitCode = errors.New("ErrUnexpectedExitCode")
	errExitCode1          = errors.New("ErrExitCode1")

	ErrDryRun       = errors.New("")
	ErrFailedToExec = errors.New("ErrFailedToExec")
	ErrPingFailed   = errors.New("ErrPingFailed")
)

// BitcoinCLI contains parameters for bitcoin-cli.
type BitcoinCLI struct {
	// Path
	binPath string
	// Connection
	btcNet      model.BTCNet
	rpcAddr     string
	rpcPort     string
	rpcUser     string
	rpcPassword string
}

// NewBitcoinCLI initializes a BitcoinCLI.
//
// Parameters:
//   - binPath sets path to bitcoin-cli. Both relative and absolute are acceptable.
//   - btcNet sets target Bitcoin network. A valid value must be set.
//   - rpcAddr sets IP address to which bitcoin-cli connects. If "" is set, default value will be used.
//   - rpcPort sets TCP port to which bitcoin-cli connects. If "" is set, default value will be used.
//   - rpcUser sets RPC username used by bitcoin-cli. If "" is set, default value will be used.
//   - rpcPassword sets RPC password used by bitcoin-cli. If "" is set, default value will be used.
//
// This function does NOT check the status of the target bitcoind.
// If necessary, call b.Ping after calling this function.
func NewBitcoinCLI(binPath string, btcNet model.BTCNet, rpcAddr, rpcPort, rpcUser, rpcPassword string) *BitcoinCLI {
	b := &BitcoinCLI{
		binPath:     binPath,
		btcNet:      btcNet,
		rpcAddr:     rpcAddr,
		rpcPort:     rpcPort,
		rpcUser:     rpcUser,
		rpcPassword: rpcPassword,
	}
	return b
}

func (b *BitcoinCLI) connArgs() []string {
	var s []string
	switch b.btcNet {
	case model.BTCMainnet:
		s = append(s, argChain+"main")
	case model.BTCTestnet3:
		s = append(s, argChain+"test")
	case model.BTCTestnet4:
		panic("not implemented")
	}
	if b.rpcAddr != "" {
		s = append(s, argRPCAddr+b.rpcAddr)
	}
	if b.rpcPort != "" {
		s = append(s, argRPCPort+b.rpcPort)
	}
	if b.rpcUser != "" {
		s = append(s, argRPCUser+b.rpcUser)
	}
	if b.rpcPassword != "" {
		s = append(s, argRPCPassword+b.rpcPassword)
	}
	return s
}

// Test use only.
var dryRun = false

func (b *BitcoinCLI) run(ctx context.Context, args []string) error {
	args = append(b.connArgs(), args...)
	if dryRun {
		return fmt.Errorf("%w%s %s", ErrDryRun, b.binPath, strings.Join(args, " "))
	}
	cmd := exec.CommandContext(ctx, b.binPath, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w (%v)", ErrFailedToExec, err)
	}
	ec := cmd.ProcessState.ExitCode()
	switch ec {
	case exitOK:
		return nil
	case exitERR:
		return errExitCode1
	default:
		return fmt.Errorf("%w (%v)", errUnexpectedExitCode, ec)
	}
}

func (b *BitcoinCLI) Ping(ctx context.Context) error {
	if err := b.run(ctx, []string{cmdPing}); err != nil {
		return err
	}
	return nil
}

func (b *BitcoinCLI) PutAnchor(ctx context.Context, a *model.Anchor) ([]byte, error) {
	panic("not implemented")
}

func (b *BitcoinCLI) GetAnchor(ctx context.Context, btctx []byte) (*model.AnchorRecord, error) {
	panic("not implemented")
}
