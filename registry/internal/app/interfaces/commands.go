package interfaces

import "context"

type ProcessNewOrder interface {
	Run(ctx context.Context) error
}
