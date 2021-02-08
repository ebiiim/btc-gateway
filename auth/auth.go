package auth

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/google/uuid"
	"gocloud.dev/docstore"
)

type Authenticator interface {
	AuthFunc(ctx context.Context, apiKey string, params map[string]string) bool
	io.Closer
}

var _ Authenticator = (*SpecialAuth)(nil)
var _ Authenticator = (*DocstoreAuth)(nil)

const (
	paramPathDomainID      = "dom"
	paramPathTransactionID = "tx"
)

// SpecialAuth is a dummy Authenticator for tests.
type SpecialAuth struct{}

// AuthFunc returns (apiKey == "12345").
func (*SpecialAuth) AuthFunc(_ context.Context, apiKey string, _ map[string]string) bool {
	return apiKey == "12345"
}

// Close returns nil.
func (*SpecialAuth) Close() error {
	return nil
}

// var AuthFn AuthFunc = specialAuthFunc

// NextKeyFunc generates next key.
type NextKeyFunc func() string

// NextKeyFn generates next key and is used by DocstoreAuth.
// Default is UUID Version 4.
var NextKeyFn = UUIDNextKey

// UUIDNextKey returns an UUID Version 4, without hyphens.
func UUIDNextKey() string {
	k, err := uuid.NewRandom()
	if err != nil {
		return ""
	}
	kb, err := k.MarshalBinary()
	if err != nil {
		return ""
	}
	return hex.EncodeToString(kb)
}

// APIKey contains API key and scope.
type APIKey struct {
	Key                 string `docstore:"key"`
	ScopeRegisterAll    bool   `docstore:"scope_register_all"`    // globalAdmin: You can do whatever you want.
	ScopeRegisterDomain bool   `docstore:"scope_register_domain"` // domainAdmin: DomainID will be checked.
	DomainID            string `docstore:"domid"`
	Note                string `docstore:"note"`
}

// Errors
var (
	ErrCouldNotOpenKeyStore  = errors.New("ErrCouldNotOpenKeyStore")
	ErrCouldNotCloseKeyStore = errors.New("ErrCouldNotCloseKeyStore")
	ErrCouldNotAuthenticate  = errors.New("ErrCouldNotAuthenticate")
	ErrCouldNotDeleteKey     = errors.New("ErrCouldNotDeleteKey")
	ErrKeyNotFound           = errors.New("ErrKeyNotFound")
	ErrCouldNotGenerateKey   = errors.New("ErrCouldNotGenerateKey")
)

// DocstoreAuth provides AuthFunc and is backed by docstore.
type DocstoreAuth struct {
	conn string
	coll *docstore.Collection
	once sync.Once
}

// AuthFunc handles authentication.
func (a *DocstoreAuth) AuthFunc(ctx context.Context, apiKey string, params map[string]string) bool {
	// TODO: check error and put logs if unexpected error
	b, _ := a.Do(ctx, apiKey, params[paramPathDomainID])
	return b
}

// Do authenticates apiKey.
func (a *DocstoreAuth) Do(ctx context.Context, apiKey string, domainID string) (bool, error) {
	k := &APIKey{
		Key: apiKey,
	}
	if err := a.coll.Get(ctx, k); err != nil {
		// NotFound: not empty but not found in docstore
		// InvalidArgument: empty string
		if strings.Contains(err.Error(), "code=NotFound") || strings.Contains(err.Error(), "code=InvalidArgument") {
			return false, nil
		}
		return false, fmt.Errorf("%w (%v)", ErrCouldNotAuthenticate, err)
	}
	if k.ScopeRegisterAll {
		return true, nil
	}
	if k.ScopeRegisterDomain && k.DomainID == domainID {
		return true, nil
	}
	return false, nil
}

// Generate generates a new APIKey and inserts it into datastore.
func (a *DocstoreAuth) Generate(ctx context.Context, domID string, isGlobalAdmin bool, note string) (*APIKey, error) {
	key := NextKeyFn()
	if key == "" {
		return nil, fmt.Errorf(`%w (NextKeyFn error)`, ErrCouldNotGenerateKey)
	}
	k := &APIKey{
		Key:                 key,
		ScopeRegisterDomain: !isGlobalAdmin,
		ScopeRegisterAll:    isGlobalAdmin,
		DomainID:            domID,
		Note:                note,
	}
	if err := a.coll.Create(ctx, k); err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrCouldNotGenerateKey, err)
	}
	return k, nil
}

// Delete deletes the APIKey specified by apiKey from datastore.
// No errors returned when the APIKey does not exist.
func (a *DocstoreAuth) Delete(ctx context.Context, apiKey string) error {
	k := &APIKey{
		Key: apiKey,
	}
	if err := a.coll.Delete(ctx, k); err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotDeleteKey, err)
	}
	return nil
}

// MustNewDocstoreAuth initializes an DocstoreAuth,
// panics if failed to access datastore.
func MustNewDocstoreAuth(conn string) *DocstoreAuth {
	a := &DocstoreAuth{
		conn: conn,
		coll: nil,
	}
	if err := a.Open(); err != nil {
		panic(fmt.Sprintf("%v conn=%s", err, conn))
	}
	return a
}

func (a *DocstoreAuth) open() error {
	coll, err := docstore.OpenCollection(context.Background(), a.conn)
	if err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotOpenKeyStore, err)
	}
	a.coll = coll
	return nil
}

// Open opens a.coll once.
func (a *DocstoreAuth) Open() error {
	var oErr error
	a.once.Do(func() { oErr = a.open() })
	if oErr != nil {
		return oErr
	}
	return nil
}

// Close closes the DocstoreAuth.
func (a *DocstoreAuth) Close() error {
	if err := a.Open(); err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotCloseKeyStore, err)
	}
	if err := a.coll.Close(); err != nil {
		return fmt.Errorf("%w (%v)", ErrCouldNotCloseKeyStore, err)
	}
	return nil
}
