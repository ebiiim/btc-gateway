package btc

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ebiiim/btc-gateway/model"
)

type cliCmd string

const (
	cmdPing                         = "ping"
	cmdGetBalance                   = "getbalance"
	cmdGetTransaction               = "gettransaction"
	cmdCreateRawTransaction         = "createrawtransaction"
	cmdSignRawTransactionWithWallet = "signrawtransactionwithwallet"
	cmdSendRawTransaction           = "sendrawtransaction"
	cmdDecodeRawTransaction         = "decoderawtransaction"
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
	exitOK              = 0
	exitERR             = 1
	exitInvalidTXID     = 5
	exitWrongSizeTXID   = 8
	exitWalletNotLoaded = 18
	exitTxDecodeFailed  = 22
	exitTxAlreadySpent  = 25
	exitTxAlreadyExists = 27
)

// Errors
var (
	ErrDryRun               = errors.New("")
	ErrFailedToExec         = errors.New("ErrFailedToExec")
	ErrFailedToDecode       = errors.New("ErrFailedToDecode")
	ErrInvalidOpReturn      = errors.New("ErrInvalidOpReturn")
	ErrInconsistentBTCNet   = errors.New("ErrInconsistentBTCNet")
	ErrUnexpectedExitCode   = errors.New("ErrUnexpectedExitCode")
	ErrExitCode1            = errors.New("ErrExitCode1")
	ErrPingFailed           = errors.New("ErrPingFailed")
	ErrInvalidTransactionID = errors.New("ErrInvalidTransactionID")
	ErrWalletNotLoaded      = errors.New("ErrWalletNotLoaded")
	ErrInvalidFee           = errors.New("ErrInvalidFee")
	ErrFailedToSign         = errors.New("ErrFailedToSign")
	ErrTxDecodeFailed       = errors.New("ErrTxDecodeFailed")
	ErrTxAlreadySpent       = errors.New("ErrTxAlreadySpent")
	ErrTxAlreadyExists      = errors.New("ErrTxAlreadyExists")
	ErrNotEnoughBalance     = errors.New("ErrNotEnoughBalance")
	ErrNotEnoughConfirm     = errors.New("ErrNotEnoughConfirm")
)

// BitcoinCLI contains parameters for bitcoin-cli.
// Parameters are read only so this struct does not have state.
//
// Excepts: xBTCAddr and xTransactionID are mutable. This is under consideration.
type BitcoinCLI struct {
	binPath     string
	btcNet      model.BTCNet
	rpcAddr     string
	rpcPort     string
	rpcUser     string
	rpcPassword string

	// Set by XPrepareAnchor and used by PutAnchor only.
	xBTCAddr       string
	xTransactionID []byte
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

func removeCRLF(buf *bytes.Buffer) *bytes.Buffer {
	var buf2 bytes.Buffer
	for {
		b, err := buf.ReadByte()
		if err == io.EOF {
			return &buf2
		}
		switch b {
		case 0x000d, 0x000a: // CR, LF
			continue
		default:
			_ = buf2.WriteByte(b) // always nil
		}
	}
}

// Test use only.
var dryRun = false

// run returns (stdout, stderr, error).
func (b *BitcoinCLI) run(ctx context.Context, args []string) (*bytes.Buffer, *bytes.Buffer, error) {
	args = append(b.connArgs(), args...)
	// log.Printf("[Trace] %s %s\n", b.binPath, strings.Join(args, " "))
	if dryRun {
		return nil, nil, fmt.Errorf("%w%s %s", ErrDryRun, b.binPath, strings.Join(args, " "))
	}
	cmd := exec.CommandContext(ctx, b.binPath, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		switch ec := cmd.ProcessState.ExitCode(); ec {
		case -1:
			return nil, nil, fmt.Errorf("%w (%v)", ErrFailedToExec, err)
		case exitOK:
			return &stdout, &stderr, nil
		case exitERR:
			return &stdout, &stderr, ErrExitCode1
		case exitInvalidTXID, exitWrongSizeTXID:
			return &stdout, &stderr, ErrInvalidTransactionID
		case exitWalletNotLoaded:
			return &stdout, &stderr, ErrWalletNotLoaded
		case exitTxDecodeFailed:
			return &stdout, &stderr, ErrTxDecodeFailed
		case exitTxAlreadySpent:
			return &stdout, &stderr, ErrTxAlreadySpent
		case exitTxAlreadyExists:
			return &stdout, &stderr, ErrTxAlreadyExists
		default:
			return &stdout, &stderr, fmt.Errorf("%w (%v)", ErrUnexpectedExitCode, ec)
		}
	}
	return &stdout, &stderr, nil
}

// Ping pings bitcoind and returns nil if successful.
//
// Possible errors: ErrExitCode1|ErrUnexpectedExitCode|ErrFailedToExec
func (b *BitcoinCLI) Ping(ctx context.Context) error {
	if stdout, stderr, err := b.run(ctx, []string{cmdPing}); err != nil {
		if errors.Is(err, ErrDryRun) {
			return err
		}
		return fmt.Errorf("%w (stdout=%s, stderr=%s)", err, stdout.String(), stderr.String())
	}
	return nil
}

// GetBalance returns balance of the default wallet.
//
// Possible errors: ErrWalletNotLoaded|ErrUnexpectedExitCode|ErrFailedToExec
func (b *BitcoinCLI) GetBalance(ctx context.Context) (string, error) {
	stdout, stderr, err := b.run(ctx, []string{cmdGetBalance})
	if err != nil {
		if errors.Is(err, ErrDryRun) {
			return "", err
		}
		return "", fmt.Errorf("%w (stdout=%s, stderr=%s)", err, stdout.String(), stderr.String())
	}
	stdout = removeCRLF(stdout)
	s := stdout.String()
	if len(s) == 0 {
		return "", fmt.Errorf("%w (%v)", ErrFailedToDecode, err)
	}
	return s, nil
}

// GetTransaction returns a transaction in JSON.
//
// Possible errors: ErrInvalidTransactionID|ErrWalletNotLoaded|ErrUnexpectedExitCode|ErrFailedToExec
func (b *BitcoinCLI) GetTransaction(ctx context.Context, txid []byte) (*bytes.Buffer, error) {
	stdout, stderr, err := b.run(ctx, []string{cmdGetTransaction, hex.EncodeToString(txid)})
	if err != nil {
		if errors.Is(err, ErrDryRun) {
			return nil, err
		}
		return nil, fmt.Errorf("%w (stdout=%s, stderr=%s)", err, stdout.String(), stderr.String())
	}
	return stdout, nil
}

// Possible errors: ErrFailedToDecode|ErrNotEnoughBalance
func calcFee(bal string, feeSatoshi uint) (string, error) {
	f64Bal, err := strconv.ParseFloat(bal, 64)
	if err != nil {
		return "", fmt.Errorf("%w (%+v)", ErrFailedToDecode, err)
	}
	var rate float64 = 100_000_000
	f64Bal = float64(int(f64Bal*rate)-int(feeSatoshi)) / rate
	if f64Bal < 0 {
		return "", fmt.Errorf("%w (%.8f)", ErrNotEnoughBalance, f64Bal)
	}
	return fmt.Sprintf("%.8f", f64Bal), err
}

// ParseTransactionReceived returns received amount of the given Bitcoin address.
// Only counts the first received amount of the given address.
// Returns error if no received.
func ParseTransactionReceived(txJSON *bytes.Buffer, recvAddr string) (string, error) {
	var val map[string]interface{}
	if err := json.NewDecoder(txJSON).Decode(&val); err != nil {
		return "", fmt.Errorf("%w (%v)", ErrFailedToDecode, err)
	}
	// Parse { ..., "details": [ { "address": "abc", "category": "receive", "amount": 0.123 } ] }
	details, ok := val["details"].([]interface{})
	if !ok {
		return "", fmt.Errorf("%w (root->details)", ErrFailedToDecode)
	}
	for idx, d := range details {
		dd, ok := d.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("%w (root->details[%d])", ErrFailedToDecode, idx)
		}
		// category == receive ? pass : continue
		cat, ok := dd["category"].(string)
		if !ok {
			return "", fmt.Errorf("%w (root->details[%d]->category)", ErrFailedToDecode, idx)
		}
		if cat != "receive" {
			continue
		}
		// address == ${recvAddr} ? pass : continue
		addr, ok := dd["address"].(string)
		if !ok {
			return "", fmt.Errorf("%w (root->details[%d]->address)", ErrFailedToDecode, idx)
		}
		if addr != recvAddr {
			continue
		}
		// amount is float ? return : ErrFailedToDecode
		amo, ok := dd["amount"].(float64)
		if !ok {
			return "", fmt.Errorf("%w (root->details[%d]->amount)", ErrFailedToDecode, idx)
		}
		return fmt.Sprintf("%.8f", amo), nil
	}
	return "", fmt.Errorf("%w (not found)", ErrFailedToDecode)
}

// CreateRawTransactionForAnchor creates a raw transaction with one vout and one OP_RETURN.
//
// Parameters:
//   - fromTxid sets UTXO.
//   - balance sets balance of fromTxid.
//   - toAddr sets destination Bitcoin address.
//   - fee sets transaction fee in Satoshi.
//     - Info: BTC 0.0002 = Satoshi 20,000
//   - data sets OP_RETURN data. Up to 80 bytes.
//
// Possible errors: ErrInvalidFee|ErrExitCode1|ErrUnexpectedExitCode|ErrFailedToExec
func (b *BitcoinCLI) CreateRawTransactionForAnchor(ctx context.Context, fromTxid []byte, balance string, toAddr string, fee uint, data []byte) ([]byte, error) {
	sFee, err := calcFee(balance, fee)
	if err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrInvalidFee, err)
	}
	argFmt0 := `[{"txid": "%s", "vout": 0}]`
	argFmt1 := `[{"%s": %s}, {"data": "%s"}]`
	arg0 := fmt.Sprintf(argFmt0, hex.EncodeToString(fromTxid))
	arg1 := fmt.Sprintf(argFmt1, toAddr, sFee, hex.EncodeToString(data))
	stdout, stderr, err := b.run(ctx, []string{cmdCreateRawTransaction, arg0, arg1})
	if err != nil {
		if errors.Is(err, ErrDryRun) {
			return nil, err
		}
		return nil, fmt.Errorf("%w (stdout=%s, stderr=%s)", err, stdout.String(), stderr.String())
	}
	stdout = removeCRLF(stdout)
	bs, err := hex.DecodeString(stdout.String())
	if err != nil || len(bs) == 0 {
		return nil, fmt.Errorf("%w (%v)", ErrFailedToDecode, err)
	}
	return bs, nil
}

// SignRawTransactionWithWallet signs the given transaction with the default wallet in bitcoin-cli and returns JSON.
//
// Possible errors: ErrWalletNotLoaded|ErrUnexpectedExitCode|ErrFailedToExec
func (b *BitcoinCLI) SignRawTransactionWithWallet(ctx context.Context, rawTx []byte) (*bytes.Buffer, error) {
	stdout, stderr, err := b.run(ctx, []string{cmdSignRawTransactionWithWallet, hex.EncodeToString(rawTx)})
	if err != nil {
		if errors.Is(err, ErrDryRun) {
			return nil, err
		}
		return nil, fmt.Errorf("%w (stdout=%s, stderr=%s)", err, stdout.String(), stderr.String())
	}
	stdout = removeCRLF(stdout)
	return stdout, nil
}

// ParseSignRawTransactionWithWallet parses the response from b.SignRawTransactionWithWallet.
func ParseSignRawTransactionWithWallet(stdout io.Reader) ([]byte, error) {
	var val map[string]interface{}
	if err := json.NewDecoder(stdout).Decode(&val); err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrFailedToDecode, err)
	}
	// Parse { "hex": "12345", "complete": true, ... }
	// return complete ? hex : ErrFailedToSign
	completed, ok := val["complete"].(bool)
	if !ok {
		return nil, fmt.Errorf("%w (root->complete)", ErrFailedToDecode)
	}
	hexStr, ok := val["hex"].(string)
	if !ok {
		return nil, fmt.Errorf("%w (root->hex)", ErrFailedToDecode)
	}
	if !completed {
		return nil, fmt.Errorf("%w (%s)", ErrFailedToSign, hexStr)
	}
	bs, err := hex.DecodeString(hexStr)
	if err != nil || len(bs) == 0 {
		return nil, fmt.Errorf("%w (%v)", ErrFailedToDecode, err)
	}
	return bs, nil
}

// SendRawTransaction sends the given signed raw transaction and returns transaction ID.
//
// Possible errors: ErrTxAlreadySpent|ErrTxAlreadyExists|ErrUnexpectedExitCode|ErrFailedToExec
func (b *BitcoinCLI) SendRawTransaction(ctx context.Context, signedRawTx []byte) ([]byte, error) {
	stdout, stderr, err := b.run(ctx, []string{cmdSendRawTransaction, hex.EncodeToString(signedRawTx)})
	if err != nil {
		if errors.Is(err, ErrDryRun) {
			return nil, err
		}
		return nil, fmt.Errorf("%w (stdout=%s, stderr=%s)", err, stdout.String(), stderr.String())
	}
	stdout = removeCRLF(stdout)
	bs, err := hex.DecodeString(stdout.String())
	if err != nil || len(bs) == 0 {
		return nil, fmt.Errorf("%w (%v)", ErrFailedToDecode, err)
	}
	return bs, nil
}

// Transaction fee in Satoshi.
const (
	feeNormal = 20_000
	feeLarge  = 30_000
	feeSmall  = 10_000
)

// txFee sets fee.
var txFee uint = feeNormal

// XPrepareAnchor sets b.xTransactionID and b.xBTCAddr.
// As b.PutAnchor does not update b.xTransactionID after sending the transaction,
// this method should be called every time after PutAnchor (or create a new BitcoinCLI instance).
func (b *BitcoinCLI) XPrepareAnchor(txid []byte, btcAddr string) {
	b.xTransactionID = txid
	b.xBTCAddr = btcAddr
}

const (
	leastConfirmationNeeded = 6
)

// PutAnchor anchors the given Anchor by sending a Bitcoin transaction and returns its transaction ID.
func (b *BitcoinCLI) PutAnchor(ctx context.Context, a *model.Anchor) ([]byte, error) {
	// Check the given Anchor.
	if a.BTCNet != b.btcNet {
		return nil, fmt.Errorf("%w (Anchor: %s, BitcoinCLI: %s) (PutAnchor)", ErrInconsistentBTCNet, a.BTCNet, b.btcNet)
	}

	// Check the bitcoind.
	err := b.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w (PutAnchor)", err)
	}

	// Get UTXO balance and check confirmations.
	fromTx, err := b.GetTransaction(ctx, b.xTransactionID)
	if err != nil {
		return nil, fmt.Errorf("%w (PutAnchor)", err)
	}
	var bufR, bufC bytes.Buffer
	w := io.MultiWriter(&bufR, &bufC)
	io.Copy(w, fromTx)
	balance, err := ParseTransactionReceived(&bufR, b.xBTCAddr)
	if err != nil {
		return nil, fmt.Errorf("%w (PutAnchor)", err)
	}
	confs, err := ParseTransactionConfirmations(&bufC)
	if err != nil {
		return nil, fmt.Errorf("%w (GetAnchor)", err)
	}
	if confs < leastConfirmationNeeded {
		return nil, fmt.Errorf("%w (confirmations=%d) (GetAnchor)", ErrNotEnoughConfirm, confs)
	}

	// Encode OP_RETURN.
	tmp := model.EncodeOpReturn(a)
	opRet := tmp[:]

	// Create, sign, and send the anchor transaction.
	rawTx, err := b.CreateRawTransactionForAnchor(ctx, b.xTransactionID, balance, b.xBTCAddr, txFee, opRet)
	if err != nil {
		return nil, fmt.Errorf("%w (PutAnchor)", err)
	}
	signedTxReader, err := b.SignRawTransactionWithWallet(ctx, rawTx)
	if err != nil {
		return nil, fmt.Errorf("%w (PutAnchor)", err)
	}
	signedTx, err := ParseSignRawTransactionWithWallet(signedTxReader)
	if err != nil {
		return nil, fmt.Errorf("%w (PutAnchor)", err)
	}
	sentTxid, err := b.SendRawTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("%w (PutAnchor)", err)
	}

	return sentTxid, nil
}

// ParseTransactionConfirmations returns confirmations of the given transaction.
func ParseTransactionConfirmations(txJSON *bytes.Buffer) (uint, error) {
	var val map[string]interface{}
	if err := json.NewDecoder(txJSON).Decode(&val); err != nil {
		return 0, fmt.Errorf("%w (%v)", ErrFailedToDecode, err)
	}
	// Parse { ..., "confirmations": 3, ... }
	confs, ok := val["confirmations"].(float64)
	if !ok {
		return 0, fmt.Errorf("%w (root->confirmations)", ErrFailedToDecode)
	}
	return uint(confs), nil
}

// ParseTransactionTime returns time of the given transaction.
func ParseTransactionTime(txJSON *bytes.Buffer) (time.Time, error) {
	var val map[string]interface{}
	if err := json.NewDecoder(txJSON).Decode(&val); err != nil {
		return time.Time{}, fmt.Errorf("%w (%v)", ErrFailedToDecode, err)
	}
	// Parse { ..., "time": 1611334493, ... }
	unixT, ok := val["time"].(float64)
	if !ok {
		return time.Time{}, fmt.Errorf("%w (root->time)", ErrFailedToDecode)
	}
	return time.Unix(int64(unixT), 0), nil
}

// ParseTransactionRawHex returns raw data of the given transaction.
func ParseTransactionRawHex(txJSON *bytes.Buffer) ([]byte, error) {
	var val map[string]interface{}
	if err := json.NewDecoder(txJSON).Decode(&val); err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrFailedToDecode, err)
	}
	// Parse { ..., "hex": "12345" }
	hexStr, ok := val["hex"].(string)
	if !ok {
		return nil, fmt.Errorf("%w (root->hex)", ErrFailedToDecode)
	}
	bs, err := hex.DecodeString(hexStr)
	if err != nil || len(bs) == 0 {
		return nil, fmt.Errorf("%w (%+v)", ErrFailedToDecode, hexStr)
	}
	return bs, nil

}

// DecodeRawTransaction returns a raw transaction in JSON.
//
// Possible errors: ErrTxDecodeFailed|ErrWalletNotLoaded|ErrUnexpectedExitCode|ErrFailedToExec
func (b *BitcoinCLI) DecodeRawTransaction(ctx context.Context, txdata []byte) (*bytes.Buffer, error) {
	stdout, stderr, err := b.run(ctx, []string{cmdDecodeRawTransaction, hex.EncodeToString(txdata)})
	if err != nil {
		if errors.Is(err, ErrDryRun) {
			return nil, err
		}
		return nil, fmt.Errorf("%w (stdout=%s, stderr=%s)", err, stdout.String(), stderr.String())
	}
	return stdout, nil
}

// ParseRawTransactionOpReturn returns OP_RETURN value of the given raw transaction.
func ParseRawTransactionOpReturn(rawTxJSON *bytes.Buffer) ([]byte, error) {
	var val map[string]interface{}
	if err := json.NewDecoder(rawTxJSON).Decode(&val); err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrFailedToDecode, err)
	}
	// Parse { ..., "vout": [ { "scriptPubKey": { "asm": "OP_RETURN 12345", ... }, ... } ] }
	vouts, ok := val["vout"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("%w (root->vout)", ErrFailedToDecode)
	}
	for idx, vout := range vouts {
		o, ok := vout.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%w (root->vout[%d])", ErrFailedToDecode, idx)
		}
		// scriptPubKey ?
		spk, ok := o["scriptPubKey"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%w (root->vout[%d]->scriptPubKey)", ErrFailedToDecode, idx)
		}
		// scriptPubKey.asm ?
		asm, ok := spk["asm"].(string)
		if !ok {
			return nil, fmt.Errorf("%w (root->vout[%d]->scriptPubKey->asm)", ErrFailedToDecode, idx)
		}
		// asm == "OP_RETURN 12345" ? pass : continue
		if !strings.HasPrefix(asm, "OP_RETURN ") {
			continue
		}
		opRet, err := hex.DecodeString(strings.TrimPrefix(asm, "OP_RETURN "))
		if err != nil {
			return nil, fmt.Errorf("%w (asm)", ErrFailedToDecode)
		}
		return opRet, nil
	}
	return nil, fmt.Errorf("%w (not found)", ErrFailedToDecode)
}

// GetAnchor returns an AnchorRecord by searching the given Bitcoin transaction ID and parsing its data.
func (b *BitcoinCLI) GetAnchor(ctx context.Context, btctx []byte) (*model.AnchorRecord, error) {
	// Check the bitcoind.
	err := b.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w (GetAnchor)", err)
	}

	// Parse the given transaction and get data.
	tx, err := b.GetTransaction(ctx, btctx)
	if err != nil {
		return nil, fmt.Errorf("%w (GetAnchor)", err)
	}
	var bufT, bufC, bufH bytes.Buffer
	w := io.MultiWriter(&bufT, &bufC, &bufH)
	io.Copy(w, tx)
	tts, err := ParseTransactionTime(&bufT)
	if err != nil {
		return nil, fmt.Errorf("%w (GetAnchor)", err)
	}
	tcs, err := ParseTransactionConfirmations(&bufC)
	if err != nil {
		return nil, fmt.Errorf("%w (GetAnchor)", err)
	}
	tHex, err := ParseTransactionRawHex(&bufH)
	if err != nil {
		return nil, fmt.Errorf("%w (GetAnchor)", err)
	}
	rawTx, err := b.DecodeRawTransaction(ctx, tHex)
	if err != nil {
		return nil, fmt.Errorf("%w (GetAnchor)", err)
	}
	opRetSlice, err := ParseRawTransactionOpReturn(rawTx)
	if err != nil {
		return nil, fmt.Errorf("%w (GetAnchor)", err)
	}

	// Decode OP_RETURN.
	if len(opRetSlice) != 80 {
		return nil, fmt.Errorf("%w (GetAnchor)", ErrInvalidOpReturn)
	}
	var opRet [80]byte
	copy(opRet[0:80], opRetSlice[0:80])
	a, err := model.DecodeOpReturn(opRet)
	if err != nil {
		return nil, fmt.Errorf("%w (%v) (GetAnchor)", ErrInvalidOpReturn, err)
	}

	// This is not a complete AnchorRecord.
	// Only data from the Bitcoin transaction is included.
	r := model.AnchorRecord{
		Anchor:           a,
		BTCTransactionID: btctx,
		TransactionTime:  tts,
		Confirmations:    tcs,
	}
	return &r, nil
}
