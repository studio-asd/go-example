package pghelper

import (
	"context"
	"testing"

	"github.com/studio-asd/pkg/postgres"

	"github.com/studio-asd/go-example/internal/testing/pghelper/testdata"
)

var _ PGQuery = (*testQuery)(nil)

type testQuery struct {
	pg *postgres.Postgres
}

func (t *testQuery) Postgres() *postgres.Postgres {
	return t.pg
}

func TestFork(t *testing.T) {
	th, err := New(context.Background(), Config{
		DatabaseName:   "test_fork",
		EmbeddedSchema: testdata.EmbeddedSchema,
	})
	if err != nil {
		t.Fatal(err)
	}

	loopNum := 10
	schemas := make(map[string]struct{})
	for range loopNum {
		newTh, err := th.ForkPostgresSchema(context.Background(), th.Postgres(), "public")
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
