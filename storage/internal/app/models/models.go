package models

type OrderStatus uint8

const (
	Pending OrderStatus = iota
	Canceled
	Completed
	Rejected
)

type TransactionType uint8

const (
	Reservation TransactionType = iota
	Cancelation
)

type StorageItem struct {
	ID        uint
	ProductID uint
	Count     uint16
}

type StorageTransaction struct {
	ID      uint
	OrderID uint
	Items   []*StorageTransactionItem
	Type    TransactionType
}

type StorageTransactionItem struct {
	ID            uint
	ProductID     uint
	TransactionID uint
	OrderID       uint
	Count         uint16
}
