package bootstrap

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"slices"

	"golang.org/x/mod/semver"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	pg "github.com/studio-asd/pkg/postgres"
	"github.com/studio-asd/pkg/srun"

	goexampledbchema "github.com/studio-asd/go-example/database/schemas/go-example"
	userdbschema "github.com/studio-asd/go-example/database/schemas/user_data"
)

type bootstrapper interface {
	Version() string
	Run(context.Context) error
	Check(context.Context) error
}

// Bootstrap service bootstraps the application by inserting the necessary data into the database.
type Bootstrap struct {
	logger        *slog.Logger
	bootstrappers []bootstrapper
}

type Params struct {
	// Currently we are passing the pg.Postgres because its easier and we don't have to create another configuration
	// only for bootstrapper.
	GoExampleDB *pg.Postgres
	UserDB      *pg.Postgres
}

func New(params Params) (*Bootstrap, error) {
	goExampleDBMigrator, err := createMigrator(params.GoExampleDB.Config(), goexampledbchema.EmbeddedSchema)
	if err != nil {
		return nil, err
	}
	userDBMigrator, err := createMigrator(params.UserDB.Config(), userdbschema.EmbeddedSchema)
	if err != nil {
		return nil, err
	}

	// Put the bootstrapper for each version here, although we will re-sort and check the bootstrapper later on, please always
	// put the new bootstrapper below the previous one.
	b := []bootstrapper{
		&v0Bootstrapper{
			goExamplePG:         params.GoExampleDB,
			userPG:              params.UserDB,
			goExampleDBMigrator: goExampleDBMigrator,
			userDBMigrator:      userDBMigrator,
		},
	}
	checkAndSortBootstrappers(b)

	return &Bootstrap{
		bootstrappers: b,
	}, nil
}

func (b *Bootstrap) Init(ctx srun.Context) error {
	b.logger = ctx.Logger
	return nil
}

type ExecuteParams struct {
	All     bool
	Version string
}

func (b *Bootstrap) Execute(ctx context.Context, params ExecuteParams) error {
	b.logger.InfoContext(ctx, "Executing bootstrap...")
	if !params.All {
		b.logger.InfoContext(ctx, "Selecting specific version for bootstrap", "bootstrap_version", params.Version)
		versionIndex := slices.IndexFunc(b.bootstrappers, func(e bootstrapper) bool {
			if e.Version() == params.Version {
				return true
			}
			return false
		})
		if versionIndex == -1 {
			return fmt.Errorf("version %s does not exists in the bootstrapper", params.Version)
		}
		boot := b.bootstrappers[versionIndex]
		if err := boot.Run(ctx); err != nil {
			return fmt.Errorf("[bootstrapper] failed to bootstrap for version %s: %v", boot.Version(), err)
		}
		if err := boot.Check(ctx); err != nil {
			return fmt.Errorf("[bootstrapper] check is failing for version %s: %v", boot.Version(), err)
		}
		return nil
	}

	b.logger.InfoContext(ctx, "Selecting all versions for bootstrap")
	for _, bs := range b.bootstrappers {
		b.logger.InfoContext(ctx, "Running bootstrap", "boostrap_version", bs.Version())
		if err := bs.Run(ctx); err != nil {
			return fmt.Errorf("[bootstrapper] failed to bootstrap for version %s: %v", bs.Version(), err)
		}
		if err := bs.Check(ctx); err != nil {
			return fmt.Errorf("[bootstrapper] check is failing for version %s: %v", bs.Version(), err)
		}
	}
	return nil
}

func createMigrator(connectConfig pg.ConnectConfig, embeddedSchema embed.FS) (*migrate.Migrate, error) {
	dsn, err := connectConfig.DSN()
	if err != nil {
		return nil, err
	}
	ioFS, err := iofs.New(embeddedSchema, "migrations")
	if err != nil {
		return nil, err
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", ioFS, dsn.URL())
	if err != nil {
		return nil, err
	}
	return migrator, nil
}

func checkAndSortBootstrappers(b []bootstrapper) error {
	s := map[string]struct{}{}
	for _, v := range b {
		_, ok := s[v.Version()]
		if !ok {
			if !semver.IsValid(v.Version()) {
				return fmt.Errorf("%s is not a valid semver", v.Version())
			}
			s[v.Version()] = struct{}{}
			continue
		}
		return fmt.Errorf("version %s already exists, please check the list of bootstrapper", v.Version())
	}

	// Ensure that we really have a sorted versions as we will tests all the bootstrapper from the lowest to highest version.
	slices.SortFunc(b, func(a, b bootstrapper) int {
		return semver.Compare(a.Version(), b.Version())
	})
	return nil
}
