package models

type (
	OrderStatus       uint8
	TransactionType   uint8
	CancelationReason uint8
)

const (
	Pending OrderStatus = iota
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
