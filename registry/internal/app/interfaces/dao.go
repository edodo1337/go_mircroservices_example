package interfaces

import (
	"context"
	"registry_service/internal/app/models"
)

type ProductPricesMap map[uint]float32

type OrdersDAO interface {
	CreateOrder(ctx context.Context, data *CreateOrderDTO) (*models.Order, error)
	DeleteOrder(ctx context.Context, orderID uint) error
	GetOrdersListByUserID(ctx context.Context, userID uint) ([]*models.Order, error)
	GetOrderByID(ctx context.Context, orderID uint) (*models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID uint, status models.OrderStatus, reasonCode uint8) (*models.Order, error)
	HealthCheck(ctx context.Context) error
}

type OrderItemsDAO interface {
	CreateOrderItemsBulk(ctx context.Context, orderID uint, items []*CreateOrderItemDTO) ([]*models.OrderItem, error)
	CreateOrderItem(ctx context.Context, orderID uint, data *CreateOrderItemDTO) (*models.OrderItem, error)
	GetOrderItemByID(ctx context.Context, orderItemID uint) (*models.OrderItem, error)
	HealthCheck(ctx context.Context) error
}

type ProductPricesDAO interface {
	GetProductPricesMap(ctx context.Context, productIDs []uint) (ProductPricesMap, error)
	HealthCheck(ctx context.Context) error
}
