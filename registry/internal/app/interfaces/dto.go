package interfaces

import "registry_service/internal/app/models"

//--------------Data Access Layer DTOs--------------

type CreateOrderDTO struct {
	UserID uint
}

type CreateOrderItemDTO struct {
	ProductID    uint
	Count        uint8
	ProductPrice float32
}

//--------------Interactors Layer DTOs--------------

type MakeOrderItemDTO struct {
	ProductID uint
	Count     uint8
}

type MakeOrderDTO struct {
	UserID     uint
	OrderItems []*MakeOrderItemDTO
}

type NewOrderItemDTO struct {
	ProductID    uint
	Count        uint8
	ProductPrice float32
}

type NewOrderDTO struct {
	UserID     uint
	OrderItems []*NewOrderItemDTO
}

//--------------Broker Layer DTOs--------------
type ServiceName uint8

const (
	Wallet ServiceName = iota
	Storage
	Registry
)

type NewOrderMsg struct {
	UserID     uint              `json:"user_id"`
	OrderID    uint              `json:"order_id"`
	OrderItems []NewOrderMsgItem `json:"order_items"`
}

type NewOrderMsgItem struct {
	OrderID      uint    `json:"order_id"`
	ProductID    uint    `json:"product_id"`
	Count        uint8   `json:"count"`
	ProductPrice float32 `json:"product_price"`
}

type OrderRejectedMsg struct {
	OrderID    uint                     `json:"order_id"`
	UserID     uint                     `json:"user_id"`
	Service    ServiceName              `json:"service"`
	ReasonCode models.CancelationReason `json:"reason_code"`
}

type OrderSuccessMsg struct {
	OrderID uint        `json:"order_id"`
	Service ServiceName `json:"service"`
}
