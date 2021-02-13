//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen -package anchor -generate "types,chi-server,spec" -include-tags "Anchor" -o anchor/api.gen.go openapi.yml

package api

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ebiiim/btcgw/api/anchor"
	"github.com/ebiiim/btcgw/auth"
	"github.com/ebiiim/btcgw/gw"
	"github.com/ebiiim/btcgw/model"

	oapimiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/getkin/kin-openapi/openapi3filter"
)

// export

var AnchorHandlerFromMux = anchor.HandlerFromMux

// TODO: cache
type GatewayService struct {
	gw.Gateway
	auth.Authenticator
}

var _ anchor.ServerInterface = (*GatewayService)(nil)

func NewGatewayService(gw gw.Gateway, authenticator auth.Authenticator) *GatewayService {
	g := &GatewayService{
		Gateway:       gw,
		Authenticator: authenticator,
	}
	return g
}

func sendGatewayServiceError(w http.ResponseWriter, code int, err error, desc string) {
	var pDesc *string = nil
	if desc != "" {
		pDesc = &desc
	}
	gsErr := anchor.Error{
		Error:            err.Error(),
		ErrorDescription: pDesc,
	}
	WriteJSON(w, code, gsErr)
}

func (g *GatewayService) GetAnchorsDomainsDomTransactionsTx(w http.ResponseWriter, r *http.Request, dom string, tx string) {
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
	WriteJSON(w, http.StatusOK, convertAnchorRecord(ar))
}

func (g *GatewayService) PatchAnchorsDomainsDomTransactionsTx(w http.ResponseWriter, r *http.Request, dom string, tx string) {
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

func (g *GatewayService) PostAnchorsDomainsDomTransactionsTx(w http.ResponseWriter, r *http.Request, dom string, tx string) {
	bdom, err1 := hex.DecodeString(dom)
	btx, err2 := hex.DecodeString(tx)
	if err1 != nil || err2 != nil {
		sendGatewayServiceError(w, http.StatusBadRequest, ErrInvalidID, ErrInvalidIDDesc)
		return
	}
	ctx := r.Context()
	_, err := g.GetRecord(ctx, bdom, btx)
	if err == nil {
		sendGatewayServiceError(w, http.StatusInternalServerError, ErrTxAlreadyExists, ErrTxAlreadyExistsDesc)
		return
	}
	btctx, err := g.RegisterTransaction(ctx, bdom, btx)
	if err != nil {
		log.Println(err)
	}
	if errors.Is(err, gw.ErrCouldNotPutAnchor) {
		sendGatewayServiceError(w, http.StatusInternalServerError, ErrRegisterFailed, ErrRegisterFailedDesc)
		return
	}
	err = g.StoreRecord(ctx, btctx)
	if err != nil {
		log.Println(err)
	}
	if errors.Is(err, gw.ErrCouldNotStoreRecord) {
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
	WriteJSON(w, http.StatusOK, convertAnchorRecord(ar))
}

// OAPIValidator sets up OpenAPI validator and must be set as a middleware.
func (g *GatewayService) OAPIValidator() func(next http.Handler) http.Handler {
	swagger, err := anchor.GetSwagger()
	if err != nil {
		panic(fmt.Sprintf("could not load swagger spec: %s", err))
	}
	// Skips validating server names.
	swagger.Servers = nil

	validatorOpts := &oapimiddleware.Options{}
	if g.Authenticator == nil {
		return oapimiddleware.OapiRequestValidatorWithOptions(swagger, validatorOpts)
	}
	validatorOpts.Options.AuthenticationFunc = func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		h := input.RequestValidationInput.Request.Header["X-Api-Key"]
		if h == nil {
			return errors.New("X-API-KEY not found")
		}
		if !g.AuthFunc(ctx, h[0], input.RequestValidationInput.PathParams) {
			return errors.New("auth failed")
		}
		return nil
	}
	return oapimiddleware.OapiRequestValidatorWithOptions(swagger, validatorOpts)
}

func (g *GatewayService) Close() error {
	err1 := g.Gateway.Close()
	err2 := g.Authenticator.Close()
	if err1 != nil || err2 != nil {
		return fmt.Errorf("%w (Gateway: %v, Authenticator: %v)", ErrCouldNotClose, err1, err2)
	}
	return nil
}

func convertAnchor(a *model.Anchor) anchor.Anchor {
	return anchor.Anchor{
		Bbc1dom: hex.EncodeToString(a.BBc1DomainID[:]),
		Bbc1tx:  hex.EncodeToString(a.BBc1TransactionID[:]),
		Chain:   a.BTCNet.String(),
		Time:    int(a.Timestamp.Unix()),
		Version: int(a.Version),
	}
}

func convertAnchorRecord(ar *model.AnchorRecord) anchor.AnchorRecord {
	var name *string = nil
	if ar.BBc1DomainName != "" {
		name = &(ar.BBc1DomainName)
	}
	var note *string = nil
	if ar.Note != "" {
		note = &(ar.Note)
	}
	return anchor.AnchorRecord{
		Anchor:        convertAnchor(ar.Anchor),
		Bbc1name:      name,
		Btctx:         hex.EncodeToString(ar.BTCTransactionID),
		Confirmations: int(ar.Confirmations),
		Note:          note,
		Time:          int(ar.TransactionTime.Unix()),
	}
}
