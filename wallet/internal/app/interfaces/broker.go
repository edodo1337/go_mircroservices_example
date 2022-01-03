package interfaces

import (
	"context"
)

type BrokerClient interface {
	GetOrderRejectedMsg(ctx context.Context) (*OrderRejectedMsg, error)
	GetNewOrderMsg(ctx context.Context) (*NewOrderMsg, error)

	SendOrderRejectedMsg(ctx context.Context, msg *OrderRejectedMsg) error
	SendPurchaseSuccess(ctx context.Context, msg *OrderSuccessMsg) error

	CloseReader() error
	CloseWriter() error

	ProduceHealthCheckMsg(ctx context.Context) error
	ConsumeHealthCheckMsg(ctx context.Context) ([]byte, error)

	HealthCheck(ctx context.Context) error
}
