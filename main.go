package main

import (
	"context"
	"os"

	"github.com/studio-asd/pkg/srun"
	"gopkg.in/yaml.v3"

	"github.com/studio-asd/pkg/resources"
	grpcserver "github.com/studio-asd/pkg/resources/grpc/server"
	pgdb "github.com/studio-asd/pkg/resources/postgres"

	"github.com/studio-asd/go-example/server"
	"github.com/studio-asd/go-example/services/bootstrap"
	ledgerapi "github.com/studio-asd/go-example/services/ledger/api"
	userapi "github.com/studio-asd/go-example/services/user/api"
)

type Config struct {
	RS resources.Config `yaml:"resources"`
}

func main() {
	srun.New(srun.Config{
		Name: "go_example",
	}).
		MustRun(run)
}

func run(ctx context.Context, runner srun.ServiceRunner) error {
	conf := Config{}
	out, err := os.ReadFile(runner.Context().Flags.Config())
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(out, &conf); err != nil {
		return err
	}
	res, err := resources.New(ctx, conf.RS)
	if err != nil {
		return err
	}

	goExamplePG := resources.MustGet[pgdb.PostgresDB](res.Container(), "go_example").Primary()
	userPG := resources.MustGet[pgdb.PostgresDB](res.Container(), "user").Primary()

	bootStrapper, err := bootstrap.New(bootstrap.Params{
		GoExampleDB: goExamplePG,
		UserDB:      userPG,
	})
	if err != nil {
		return err
	}
	if err := bootStrapper.Upgrade(ctx, bootstrap.ExecuteParams{All: true}); err != nil {
		return err
	}

	ledgerAPI := ledgerapi.New(goExamplePG)
	userAPI := userapi.New(userPG)
	grpcServer := resources.MustGet[*grpcserver.GRPCServer](res.Container(), "main")

	svc := server.New(ledgerAPI, userAPI)
	svc.RegisterAPIServices(grpcServer)

	return runner.Register(
		srun.RegisterInitServices(
			ledgerAPI,
			userAPI,
		),
		srun.RegisterRunnerServices(res),
	)
}
