package service

import (
	ledgerpg "github.com/albertwidi/go-example/ledger/internal/postgres"
)

type Ledger struct {
	q *ledgerpg.Queries
}

func New(q *ledgerpg.Queries) *Ledger {
	l := &Ledger{
		q: q,
	}
	return l
}
