package models

import "time"

type (
	OrderStatus       uint8
	TransactionType   uint8
	CancelationReason uint8
)

const (
	Pending OrderStatus = iota
	Paid
	Reserved
	Canceled
	Completed
	Rejected
)

const (
	Purchase TransactionType = iota
	Cancelation
)

const (
	OK CancelationReason = iota
	NotEnoughMoney
	OutOfStock
	InternalError
)

type Order struct {
	ID             uint
	UserID         uint
	CreatedAt      time.Time
	Status         OrderStatus
	OrderItems     []*OrderItem
	RejectedReason CancelationReason
}

type OrderItem struct {
	ID           uint
	OrderID      uint
	ProductID    uint
	Count        uint8
	ProductPrice float32 // TODO consider as decimal
}

type Product struct {
	ID    uint
	Title string
	Price float32 // TODO consider as decimal
}
