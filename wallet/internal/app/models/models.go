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
	Purchase TransactionType = iota
	Cancelation
)

type Wallet struct {
	ID      uint
	UserID  uint
	Balance float32
}

type WalletTransaction struct {
	ID       uint
	WalletID uint
	OrderID  uint
	Cost     float32
	Wallet   *Wallet
	Type     TransactionType
}
