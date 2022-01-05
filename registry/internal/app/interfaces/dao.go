package interfaces

import (
	"context"
	"registry_service/internal/app/models"
)

type ProductPricesMap map[uint]float32

type OrdersDAO interface {
	Create(ctx context.Context, data *CreateOrderDTO) (*models.Order, error)
	Delete(ctx context.Context, orderID uint) error
	GetListByUserID(ctx context.Context, userID uint) ([]*models.Order, error)
	GetByID(ctx context.Context, orderID uint) (*models.Order, error)
	UpdateStatus(ctx context.Context, orderID uint, status models.OrderStatus, reasonCode models.CancelationReason) (*models.Order, error)
	HealthCheck(ctx context.Context) error
	Close()
}

type OrderItemsDAO interface {
	CreateBulk(ctx context.Context, orderID uint, items []*CreateOrderItemDTO) ([]*models.OrderItem, error)
	Create(ctx context.Context, orderID uint, data *CreateOrderItemDTO) (*models.OrderItem, error)
	GetByID(ctx context.Context, orderItemID uint) (*models.OrderItem, error)
	HealthCheck(ctx context.Context) error
	Close()
}

type ProductPricesDAO interface {
	GetMap(ctx context.Context, productIDs []uint) (ProductPricesMap, error)
	GetList(ctx context.Context) ([]*models.Product, error)
	HealthCheck(ctx context.Context) error
	Close()
}
