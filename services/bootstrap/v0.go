package main

import "context"

type v0Bootstrapper struct{}

func (b *v0Bootstrapper) Run(ctx context.Context) error {
	return nil
}

func (b *v0Bootstrapper) Check(ctx context.Context) error {
	return nil
}
