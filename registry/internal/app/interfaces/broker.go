package interfaces

import (
	"context"
)

type BrokerClient interface {
	SendNewOrderMsg(ctx context.Context, msg *NewOrderMsg) error
	GetOrderRejectedMsg(ctx context.Context) (*OrderRejectedMsg, error)

	CloseReader() error
	CloseWriter() error

	ProduceHealthCheckMsg(ctx context.Context) error
	ConsumeHealthCheckMsg(ctx context.Context) ([]byte, error)

	HealthCheck(ctx context.Context) error
}
