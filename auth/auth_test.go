package auth_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ebiiim/btcgw/auth"

	_ "gocloud.dev/docstore/memdocstore"
)

func TestUUIDNextKey(t *testing.T) {
	k := auth.UUIDNextKey()
	if len(k) != 32 {
		t.Error("len(k) should be 32")
	}
}

func TestSpecialAuth(t *testing.T) {
	a := &auth.SpecialAuth{}
	if !(a.AuthFunc(nil, "12345", nil)) && a.AuthFunc(nil, "1234", nil) {
		t.Error("AuthFunc")
		t.Skip()
	}
	if err := a.Close(); err != nil {
		t.Error("Close")
		t.Skip()
	}
}

//
// Do not parallelize DocstoreAuth tests because memdocstore is NOT thread-safe.
//

func dummyNextKey(key string) auth.NextKeyFunc {
	return func() string { return key }
}

// Assumes tests will be run from package root.
var (
	testdb1 = "testdata/apikeys1.db"
	conn1   = "mem://auth_test_do/key?filename=" + testdb1
	conn2   = "mem://auth_test/key"
)

var (
	key1    = "0123456789abcdef0123456789abcdef"
	key2    = "123456789abcdef0123456789abcdef0"
	key3    = "23456789abcdef0123456789abcdef01"
	dom1    = "0123456780abcdef0120456789abc0ef"
	dom2    = "1234567890bcdef0123056789abcd0f0"
	dom3    = "23456789a0cdef0123406789abcde001"
	apikey1 = &auth.APIKey{
		Key:                 key1,
		ScopeRegisterAll:    true,
		ScopeRegisterDomain: false,
		DomainID:            dom1,
		Note:                "globalAdmin",
	}
	apikey2 = &auth.APIKey{
		Key:                 key2,
		ScopeRegisterAll:    false,
		ScopeRegisterDomain: true,
		DomainID:            dom2,
		Note:                "",
	}
	apikey3 = &auth.APIKey{
		Key:                 key3,
		ScopeRegisterAll:    false,
		ScopeRegisterDomain: true,
		DomainID:            dom3,
		Note:                "",
	}
)

func TestDocstoreAuth_Do(t *testing.T) {
	cases := []struct {
		name      string
		conn      string
		key       string
		trydomain string
		want      bool
	}{
		{"admin_self", conn1, key1, dom1, true},
		{"admin_other_1", conn1, key1, dom2, true},
		{"admin_other_2", conn1, key1, dom3, true},
		{"user_self", conn1, key2, dom2, true},
		{"user_other_1", conn1, key2, dom3, false},
		{"user_other_2", conn1, key2, dom1, false},
		{"notfound_1", conn1, "12121212121212121212121212121212", dom1, false},
		{"notfound_2", conn1, "1", dom1, false},
		{"empty", conn1, "", dom1, false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			a := auth.MustNewDocstoreAuth(c.conn)
			defer a.Close()
			ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFunc()

			got, err := a.Do(ctx, c.key, c.trydomain)
			if err != nil {
				t.Error(err)
				t.Skip()
			}
			if got != c.want {
				t.Errorf("got %v but want %v", got, c.want)
				t.Skip()
			}
		})
	}
}

func TestDocstoreAuth_Generate(t *testing.T) {
	cases := []struct {
		name    string
		conn    string
		key     string
		dom     string
		isAdmin bool
		note    string
		want    *auth.APIKey
	}{
		{"admin", conn2, key1, dom1, true, "globalAdmin", apikey1},
		{"user1", conn2, key2, dom2, false, "", apikey2},
		{"user2", conn2, key3, dom3, false, "", apikey3},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			a := auth.MustNewDocstoreAuth(c.conn)
			defer a.Close()
			ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFunc()

			auth.NextKeyFn = dummyNextKey(c.key)
			got, err := a.Generate(ctx, c.dom, c.isAdmin, c.note)
			if err != nil {
				t.Error(err)
				t.Skip()
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %+v but want %+v", got, c.want)
			}
		})
	}
}

func TestDocstoreAuth_Generate_Error(t *testing.T) {
	cases := []struct {
		name    string
		conn    string
		key     string
		dom     string
		isAdmin bool
		note    string
		want    error
	}{
		{"dup", conn2, key1, dom1, true, "globalAdmin", auth.ErrCouldNotGenerateKey},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			a := auth.MustNewDocstoreAuth(c.conn)
			defer a.Close()
			ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFunc()

			auth.NextKeyFn = dummyNextKey(c.key)
			_, err := a.Generate(ctx, c.dom, c.isAdmin, c.note)
			if err != nil {
				t.Error(err)
				t.Skip()
			}
			_, err = a.Generate(ctx, c.dom, c.isAdmin, c.note)
			if !errors.Is(err, c.want) {
				t.Errorf("got %+v but want %+v", err, c.want)
			}
		})
	}
}

func TestDocstoreAuth_Delete(t *testing.T) {
	cases := []struct {
		name    string
		conn    string
		key     string
		dom     string
		isAdmin bool
		note    string
	}{
		{"admin", conn2, key1, dom1, true, "globalAdmin"},
		{"user1", conn2, key2, dom2, false, ""},
		{"user2", conn2, key3, dom3, false, ""},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			a := auth.MustNewDocstoreAuth(c.conn)
			defer a.Close()
			ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFunc()

			auth.NextKeyFn = dummyNextKey(c.key)
			got, err := a.Generate(ctx, c.dom, c.isAdmin, c.note)
			if err != nil {
				t.Error(err)
				t.Skip()
			}
			err = a.Delete(ctx, got.Key)
			if err != nil {
				t.Error(err)
			}
			err = a.Delete(ctx, got.Key)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestDocstoreAuth_Delete_NotFound(t *testing.T) {
	cases := []struct {
		name string
		conn string
		key  string
	}{
		{"not_found", conn2, "12345"},
		{"empty", conn2, ""},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			a := auth.MustNewDocstoreAuth(c.conn)
			defer a.Close()
			ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFunc()

			err := a.Delete(ctx, c.key)
			if err != nil {
				t.Error(err)
			}
			err = a.Delete(ctx, c.key)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
