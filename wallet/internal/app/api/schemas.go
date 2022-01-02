package api

type ErrResponseMsg struct {
	Message string `json:"message"`
}

type CreateOrderRequest struct {
	UserID     uint                     `json:"user_id" validate:"min=1"`
	OrderItems []CreateOrderRequestItem `json:"order_items" validate:"min=1"`
}

type CreateOrderRequestItem struct {
	ProductID uint  `json:"product_id" validate:"min=1"`
	Count     uint8 `json:"count" validate:"min=1"`
}

type CreateOrderResponse struct {
	Status string `json:"status"`
}

type HealthCheckResposne struct {
	OrdersConn        string `json:"orders_conn"`
	OrderItemsConn    string `json:"order_items_conn"`
	ProductPricesConn string `json:"product_prices_conn"`
	BrokerConn        string `json:"broker_conn"`
}
