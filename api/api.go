//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen -package api -generate "types,chi-server,spec" -o api.gen.go openapi.yml

package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ebiiim/btc-gateway/gw"
	"github.com/ebiiim/btc-gateway/model"
)

var PrettifyResponseJSON = false

func writeJSON(w io.Writer, v interface{}) {
	e := json.NewEncoder(w)
	if PrettifyResponseJSON {
		e.SetIndent("", "  ")
	}
	e.Encode(v)
}

var (
	ErrInvalidID     = errors.New("btcgw::invalid_id")
	ErrInvalidIDDesc = "ID should be a 32 bytes binary in hexadecimal string."

	ErrTxNotFound     = errors.New("btcgw::tx_not_found")
	ErrTxNotFoundDesc = "Transaction not found."

	ErrRegisterFailed     = errors.New("btcgw::register_failed")
	ErrRegisterFailedDesc = "Could not register. There may be a system error."
)

// TODO: shutdown
// TODO: cache
type GatewayService struct {
	gw.Gateway
	ServerInterface
}

func NewGatewayService(gw gw.Gateway) *GatewayService {
	g := &GatewayService{
		Gateway: gw,
	}
	return g
}

func sendGatewayServiceError(w http.ResponseWriter, code int, err error, desc string) {
	var pDesc *string = nil
	if desc != "" {
		pDesc = &desc
	}
	gsErr := Error{
		Error:            err.Error(),
		ErrorDescription: pDesc,
	}
	w.WriteHeader(code)
	writeJSON(w, gsErr)
}

func (g *GatewayService) GetDomainsDomTransactionsTx(w http.ResponseWriter, r *http.Request, dom string, tx string) {
	bdom, err1 := hex.DecodeString(dom)
	btx, err2 := hex.DecodeString(tx)
	if err1 != nil || err2 != nil {
		sendGatewayServiceError(w, http.StatusBadRequest, ErrInvalidID, ErrInvalidIDDesc)
		return
	}
	ctx := r.Context()
	ar, err := g.GetRecord(ctx, bdom, btx)
	if err != nil {
		log.Println(err)
	}
	if errors.Is(err, gw.ErrCouldNotGetRecord) {
		sendGatewayServiceError(w, http.StatusNotFound, ErrTxNotFound, ErrTxNotFoundDesc)
		return
	}
	w.WriteHeader(http.StatusOK)
	writeJSON(w, convertAnchorRecord(ar))
}

func (g *GatewayService) PatchDomainsDomTransactionsTx(w http.ResponseWriter, r *http.Request, dom string, tx string) {
	bdom, err1 := hex.DecodeString(dom)
	btx, err2 := hex.DecodeString(tx)
	if err1 != nil || err2 != nil {
		sendGatewayServiceError(w, http.StatusBadRequest, ErrInvalidID, ErrInvalidIDDesc)
		return
	}
	ctx := r.Context()
	note := fmt.Sprintf("Updated at %s", time.Now().Format(time.RFC3339))
	err := g.RefreshRecord(ctx, bdom, btx, nil, &note)
	if err != nil {
		log.Println(err)
	}
	if errors.Is(err, gw.ErrCouldNotRefreshRecord) {
		sendGatewayServiceError(w, http.StatusNotFound, ErrTxNotFound, ErrTxNotFoundDesc)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (g *GatewayService) PostDomainsDomTransactionsTx(w http.ResponseWriter, r *http.Request, dom string, tx string) {
	bdom, err1 := hex.DecodeString(dom)
	btx, err2 := hex.DecodeString(tx)
	if err1 != nil || err2 != nil {
		sendGatewayServiceError(w, http.StatusBadRequest, ErrInvalidID, ErrInvalidIDDesc)
		return
	}
	ctx := r.Context()
	_, err := g.RegisterTransaction(ctx, bdom, btx)
	if err != nil {
		log.Println(err)
	}
	if errors.Is(err, gw.ErrCouldNotPutAnchor) {
		sendGatewayServiceError(w, http.StatusInternalServerError, ErrRegisterFailed, ErrRegisterFailedDesc)
		return
	}
	ar, err := g.GetRecord(ctx, bdom, btx)
	if err != nil {
		log.Println(err)
	}
	if errors.Is(err, gw.ErrCouldNotGetRecord) {
		sendGatewayServiceError(w, http.StatusInternalServerError, ErrRegisterFailed, ErrRegisterFailedDesc)
		return
	}
	w.WriteHeader(http.StatusOK)
	writeJSON(w, convertAnchorRecord(ar))
}

func convertAnchor(a *model.Anchor) Anchor {
	return Anchor{
		Bbc1dom: hex.EncodeToString(a.BBc1DomainID[:]),
		Bbc1tx:  hex.EncodeToString(a.BBc1TransactionID[:]),
		Chain:   a.BTCNet.String(),
		Time:    int(a.Timestamp.Unix()),
		Version: int(a.Version),
	}
}

func convertAnchorRecord(ar *model.AnchorRecord) AnchorRecord {
	var name *string = nil
	if ar.BBc1DomainName != "" {
		name = &(ar.BBc1DomainName)
	}
	var note *string = nil
	if ar.Note != "" {
		note = &(ar.Note)
	}
	return AnchorRecord{
		Anchor:        convertAnchor(ar.Anchor),
		Bbc1name:      name,
		Btctx:         hex.EncodeToString(ar.BTCTransactionID),
		Confirmations: int(ar.Confirmations),
		Note:          note,
		Time:          int(ar.TransactionTime.Unix()),
	}
}
