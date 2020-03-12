package epay

import "context"

// EnvironmentStore is an interface used for retrieving of the environment.
type EnvironmentStore interface {

	// Get gets the environment configuration of the provided name.
	Get(ctx context.Context, name string) (*Environment, error)
}
