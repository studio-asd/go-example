package bootstrap

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/studio-asd/pkg/postgres"

	"github.com/studio-asd/go-example/internal/testing/pghelper"
)

var _ bootstrapper = (*dummyBootstrapper)(nil)

var (
	goExampleDB *postgres.Postgres
	userDB      *postgres.Postgres
)

func TestMain(m *testing.M) {
	flag.Parse()
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (int, error) {
	if !testing.Short() {
		goExampleHelper, err := pghelper.New(context.Background(), pghelper.Config{
			DatabaseName: "go_example",
			SkipPrepare:  true,
		})
		if err != nil {
			return 1, err
		}
		goExampleDB = goExampleHelper.Postgres()

		userHelper, err := pghelper.New(context.Background(), pghelper.Config{
			DatabaseName: "user_data",
			SkipPrepare:  true,
		})
		if err != nil {
			return 1, err
		}
		userDB = userHelper.Postgres()
	}
	return m.Run(), nil
}

func TestBootstrapFromZero(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip()
	}

	b, err := New(Params{
		GoExampleDB: goExampleDB,
		UserDB:      userDB,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Upgrade(t.Context(), ExecuteParams{All: true}); err != nil {
		t.Fatal(err)
	}
}

func TestUpgrade(t *testing.T) {
	t.Parallel()
}

func TestCheckAndSortBootstrappers(t *testing.T) {
	t.Parallel()

	bs := []bootstrapper{
		&dummyBootstrapper{V: "v0.1"},
		&dummyBootstrapper{V: "v0.1.1"},
		&dummyBootstrapper{V: "v0.1.2"},
		&dummyBootstrapper{V: "v0.2"},
		&dummyBootstrapper{V: "v0.2.1"},
		&dummyBootstrapper{V: "v0.2.2"},
		&dummyBootstrapper{V: "v1.0"},
		&dummyBootstrapper{V: "v2.0"},
		&dummyBootstrapper{V: "v2.1"},
	}

	tests := []struct {
		name   string
		b      []bootstrapper
		expect []bootstrapper
	}{
		{
			name:   "sorted, no error",
			b:      bs,
			expect: bs,
		},
		{
			name:   "unordered, no error",
			b:      append(bs[2:], bs[:2]...),
			expect: bs,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkAndSortBootstrappers(test.b)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.expect, test.b); diff != "" {
				t.Fatalf("(-want/+got)\n%s", diff)
			}
		})
	}
}

type dummyBootstrapper struct {
	V string
}

func (b *dummyBootstrapper) Version() string {
	return b.V
}

func (b *dummyBootstrapper) Upgrade(ctx context.Context) error {
	return nil
}

func (b *dummyBootstrapper) CheckUpgrade(ctx context.Context) error {
	return nil
}

func (b *dummyBootstrapper) Rollback(ctx context.Context) error {
	return nil
}

func (b *dummyBootstrapper) CheckRollback(ctx context.Context) error {
	return nil
}
