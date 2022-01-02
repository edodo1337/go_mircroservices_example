package models

import "time"

type OrderStatus uint8

const (
	Pending OrderStatus = iota
	Canceled
	Completed
	Rejected
)

type Order struct {
	ID             uint
	UserID         uint
	CreatedAt      time.Time
	Status         OrderStatus
	OrderItems     []*OrderItem
	RejectedReason uint8
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
