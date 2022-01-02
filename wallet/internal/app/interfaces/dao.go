package interfaces

import (
	"context"
	"wallet_service/internal/app/models"
)

type WalletsDAO interface {
	GetByUserID(ctx context.Context, userID uint) (*models.Wallet, error)
	UpdateBalance(ctx context.Context, wallet *models.Wallet) (*models.Wallet, error)
	HealthCheck(ctx context.Context) error
}

type WalletTransactionsDAO interface {
	GetByOrderID(ctx context.Context, orderID uint) (*models.WalletTransaction, error)
	Create(ctx context.Context, trans *CreateWalletTransactionDTO) (*models.WalletTransaction, error)
	HealthCheck(ctx context.Context) error
}
