package interfaces

import "errors"

var (
	ErrEmptyOrderItems         = errors.New("got empty order items list")
	ErrEmptyProductIDs         = errors.New("got empty product ids list")
	ErrProductNotFound         = errors.New("product not found")
	ErrNewOrderTimeout         = errors.New("new order channel send timeout")
	ErrOrderNotFound           = errors.New("order not found")
	ErrInvalidBrokerConnParams = errors.New("invalid broker client params")
	ErrBrokerConnClosed        = errors.New("broker connection closed")
)
