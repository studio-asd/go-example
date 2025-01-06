package main

import (
	"context"

	"github.com/studio-asd/pkg/srun"
)

func main() {
	srun.New(srun.Config{}).
		MustRun(run)
}

func run(ctx context.Context, runner srun.ServiceRunner) error {
	return nil
}
