package interfaces

import "storage_service/internal/app/models"

//--------------Data Access Layer DTOs--------------

type CreateStorageTransactionDTO struct {
	OrderID uint
	Items   []*CreateStorageTransactionItemDTO
	Type    models.TransactionType
}

type CreateStorageTransactionItemDTO struct {
	ProductID uint
	Count     uint16
}

//--------------Interactors Layer DTOs--------------

type OrderItemDTO struct {
	ProductID    uint
	Count        uint16
	ProductPrice float32
}

type OrderDTO struct {
	OrderID    uint
	UserID     uint
	OrderItems []*OrderItemDTO
}

type CancelOrderDTO struct {
	OrderID uint
	UserID  uint
	Cost    float32
}

type Transaction struct {
	OrderID uint
	Items   []*TransactionItem
	Type    models.TransactionType
}

type TransactionItem struct {
	ProductID uint
	Count     uint16
}

//--------------Broker Layer DTOs--------------
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
	OrderID    uint    `json:"order_id"`
	UserID     uint    `json:"user_id"`
	Cost       float32 `json:"cost"`
	ReasonCode uint8   `json:"reason_code"`
}
