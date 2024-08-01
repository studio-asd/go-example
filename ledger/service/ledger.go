package service

import (
	"github.com/albertwidi/pkg/postgres"

	ledgerpg "github.com/albertwidi/go-example/ledger/postgres"
)

type Ledger struct {
	q *ledgerpg.Queries
}

func New(pg *postgres.Postgres) *Ledger {
	l := &Ledger{
		q: ledgerpg.New(pg),
	}
	return l
}
