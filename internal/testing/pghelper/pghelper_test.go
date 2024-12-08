package pghelper

import (
	"context"
	"testing"

	"github.com/albertwidi/pkg/postgres"
)

var _ PGQuery = (*testQuery)(nil)

type testQuery struct {
	pg *postgres.Postgres
}

func newTestQuery(pg *postgres.Postgres) *testQuery {
	return &testQuery{pg: pg}
}

func (t *testQuery) Postgres() *postgres.Postgres {
	return t.pg
}

func TestFork(t *testing.T) {
	th, err := New(context.Background(), "test_fork", newTestQuery)
	if err != nil {
		t.Fatal(err)
	}

	var loopNum = 10
	var schemas = make(map[string]struct{})
	for i := 0; i < loopNum; i++ {
		newTh, err := th.ForkPostgresSchema(context.Background(), th.Queries())
		if err != nil {
			t.Fatal(err)
		}
		schemas[newTh.DefaultSearchPath()] = struct{}{}
	}
	if len(schemas) != loopNum {
		t.Fatalf("expecting %d number of schemas but got %d", loopNum, len(schemas))
	}
	if err := th.Close(); err != nil {
		t.Fatal(err)
	}
	if !th.closed {
		t.Fatal("test helper is not closed")
	}
}
