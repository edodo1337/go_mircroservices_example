package interfaces

import "wallet_service/internal/app/models"

//--------------Data Access Layer DTOs--------------

type CreateWalletTransactionDTO struct {
	WalletID uint
	OrderID  uint
	Cost     float32
	Type     models.TransactionType
}

//--------------Interactors Layer DTOs--------------

type OrderItemDTO struct {
	ProductID    uint
	Count        uint8
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
}

type Transaction struct {
	Cost    float32
	OrderID uint
	Wallet  *models.Wallet
	Type    models.TransactionType
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
