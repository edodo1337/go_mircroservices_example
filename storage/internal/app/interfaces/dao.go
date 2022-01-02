package interfaces

import (
	"context"
	"storage_service/internal/app/models"
)

type StorageItemsDAO interface {
	GetListByProductIDs(ctx context.Context, prodIDs []uint) ([]*models.StorageItem, error)
	UpdateCountBulk(ctx context.Context, items []*models.StorageItem) error
	HealthCheck(ctx context.Context) error
}

type StorageTransactionsDAO interface {
	GetByOrderID(ctx context.Context, orderID uint) (*models.StorageTransaction, error)
	GetItemsByOrderID(ctx context.Context, orderID uint) ([]*models.StorageTransactionItem, error)
	Create(ctx context.Context, trans *CreateStorageTransactionDTO) (*models.StorageTransaction, error)
	HealthCheck(ctx context.Context) error
}
