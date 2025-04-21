package main

import (
	"context"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/studio-asd/pkg/resources"
	"github.com/studio-asd/pkg/srun"

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

	goExamplePG := res.Container().Postgres().MustGetPostgres("go_example").Primary()
	userPG := res.Container().Postgres().MustGetPostgres("user").Primary()

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
	grpcServer := res.Container().GRPC().Server.MustGetServer("main")

	svc := server.New(ledgerAPI, userAPI)
	svc.RegisterAPIServices(grpcServer)

	return runner.Register(
		srun.RegisterInitAwareServices(
			ledgerAPI,
			userAPI,
		),
		srun.RegisterRunnerAwareServices(res),
	)
}
