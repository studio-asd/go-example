package api

import (
	"context"
	"testing"

	ledgerv1 "github.com/albertwidi/go-example/proto/api/ledger/v1"
)

func TestTransact(t *testing.T) {
	a := New()
	_, err := a.Transact(context.Background(), &ledgerv1.TransactRequest{})
	if err != nil {
		t.Fatal(err)
	}
}
