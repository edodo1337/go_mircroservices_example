package interfaces

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
