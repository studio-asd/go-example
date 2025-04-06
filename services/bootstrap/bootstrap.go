package main

import "context"

type bootstrapper interface {
	Run(context.Context) error
	Check(context.Context) error
}

// Bootstrap service bootstraps the application by inserting the necessary data into the database.
type Bootstrap struct {
	versions map[string]bootstrapper
}

func New() *Bootstrap {
	return &Bootstrap{
		versions: map[string]bootstrapper{
			"v0": &v0Bootstrapper{},
		},
	}
}
