package bootstrap

import (
	"context"

	"github.com/golang-migrate/migrate/v4"

	"github.com/studio-asd/pkg/postgres"
)

type v0Bootstrapper struct {
	goExamplePG         *postgres.Postgres
	userPG              *postgres.Postgres
	goExampleDBMigrator *migrate.Migrate
	userDBMigrator      *migrate.Migrate
}

func (b *v0Bootstrapper) Version() string {
	return "v0.1"
}

func (b *v0Bootstrapper) Run(ctx context.Context) error {
	if err := b.goExampleDBMigrator.Migrate(1); err != nil {
		return err
	}
	if err := b.userDBMigrator.Migrate(1); err != nil {
		return err
	}
	return nil
}

func (b *v0Bootstrapper) Check(ctx context.Context) error {
	if err := PostgresCheckMigrations(ctx, b.goExamplePG, []int64{1}); err != nil {
		return err
	}
	if err := PostgresCheckMigrations(ctx, b.userPG, []int64{1}); err != nil {
		return err
	}
	return nil
}
