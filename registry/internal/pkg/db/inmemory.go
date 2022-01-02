package db

import (
	"context"
	in "registry_service/internal/app/interfaces"
	"registry_service/internal/app/models"
)

// ------------------------------OrdersDAO------------------------------

type InMemoryOrdersDAO struct {
	OrdersKVStore map[uint]*models.Order
	lastOrderID   uint
}

func (dao *InMemoryOrdersDAO) Create(ctx context.Context, data *in.CreateOrderDTO) (*models.Order, error) {
	dao.lastOrderID++

	order := &models.Order{
		ID:     dao.lastOrderID,
		UserID: data.UserID,
	}

	dao.OrdersKVStore[dao.lastOrderID] = order

	return order, nil
}

func (dao *InMemoryOrdersDAO) GetListByUserID(ctx context.Context, userID uint) ([]*models.Order, error) {
	panic("not implemented")
}

func (dao *InMemoryOrdersDAO) GetByID(ctx context.Context, orderID uint) (*models.Order, error) {
	panic("not implemented")
}

func (dao *InMemoryOrdersDAO) Delete(ctx context.Context, orderID uint) error {
	_, ok := dao.OrdersKVStore[orderID]
	if !ok {
		return in.ErrOrderNotFound
	}

	delete(dao.OrdersKVStore, orderID)

	return nil
}

func (dao *InMemoryOrdersDAO) UpdateStatus(
	ctx context.Context,
	orderID uint,
	status models.OrderStatus,
	reasonCode models.CancelationReason,
) (*models.Order, error) {
	order, exists := dao.OrdersKVStore[orderID]
	if !exists {
		return nil, in.ErrOrderNotFound
	}

	order.Status = status
	order.RejectedReason = reasonCode

	return order, nil
}

func (dao *InMemoryOrdersDAO) HealthCheck(ctx context.Context) error {
	return nil
}

func NewInMemoryOrdersDAO() *InMemoryOrdersDAO {
	return &InMemoryOrdersDAO{
		OrdersKVStore: make(map[uint]*models.Order),
		lastOrderID:   0,
	}
}

// ---------------------------- OrderItemsDAO----------------------------

type InMemoryOrderItemsDAO struct {
	OrderItemsKVStore map[uint]*models.OrderItem
	lastOrderItemID   uint
}

func (dao *InMemoryOrderItemsDAO) CreateBulk(
	ctx context.Context,
	orderID uint,
	items []*in.CreateOrderItemDTO,
) ([]*models.OrderItem, error) {
	if len(items) == 0 {
		return nil, in.ErrEmptyOrderItems
	}

	orderItems := make([]*models.OrderItem, 0, 10)

	for _, item := range items {
		dao.lastOrderItemID++

		orderItem := &models.OrderItem{
			ID:           dao.lastOrderItemID,
			OrderID:      orderID,
			ProductID:    item.ProductID,
			Count:        item.Count,
			ProductPrice: item.ProductPrice,
		}

		dao.OrderItemsKVStore[dao.lastOrderItemID] = orderItem

		orderItems = append(orderItems, orderItem)
	}

	return orderItems, nil
}

func (dao *InMemoryOrderItemsDAO) Create(
	ctx context.Context,
	orderID uint,
	data *in.CreateOrderItemDTO,
) (*models.OrderItem, error) {
	panic("not implemented")
}

func (dao *InMemoryOrderItemsDAO) GetByID(ctx context.Context, orderItemID uint) (*models.OrderItem, error) {
	panic("not implemented")
}

func (dao *InMemoryOrderItemsDAO) HealthCheck(ctx context.Context) error {
	return nil
}

func NewInMemoryOrderItemsDAO() *InMemoryOrderItemsDAO {
	return &InMemoryOrderItemsDAO{
		OrderItemsKVStore: make(map[uint]*models.OrderItem),
		lastOrderItemID:   0,
	}
}

// ---------------------------- ProductPricesDAO----------------------------

type InMemoryProductPricesDAO struct {
	ProductPricesKVStore map[uint]float32
}

func (dao *InMemoryProductPricesDAO) GetMap(ctx context.Context, productIDs []uint) (in.ProductPricesMap, error) {
	if len(productIDs) == 0 {
		return nil, in.ErrEmptyProductIDs
	}

	pricesMap := make(map[uint]float32)

	for _, id := range productIDs {
		val, exists := dao.ProductPricesKVStore[id]
		if !exists {
			return nil, in.ErrProductNotFound
		}

		pricesMap[id] = val
	}

	return pricesMap, nil
}

func (dao *InMemoryProductPricesDAO) HealthCheck(ctx context.Context) error {
	return nil
}

func NewInMemoryProductPricesDAO() *InMemoryProductPricesDAO {
	productPrices := map[uint]float32{
		1: 1.0,
		2: 2.0,
		3: 3.0,
		4: 4.0,
	}

	return &InMemoryProductPricesDAO{
		ProductPricesKVStore: productPrices,
	}
}
