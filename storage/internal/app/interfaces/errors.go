package interfaces

import "errors"

var (
	ErrNotEnoughMoney             = errors.New("not enough money")
	ErrNewTransactionTimeoutError = errors.New("new trans timeout error")
	ErrInvalidBrokerConnParams    = errors.New("invalid broker client params")
	ErrBrokerConnClosed           = errors.New("broker connection closed")
	ErrProductNotFoundByID        = errors.New("product not found by id")
	ErrOutOfStock                 = errors.New("product out of stock")
)
