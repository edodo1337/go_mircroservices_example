package logic

import (
	in "storage_service/internal/app/interfaces"
	"storage_service/internal/app/models"
)

func calcOrderSum(orderData *in.OrderDTO) float32 {
	var sum float32
	for _, v := range orderData.OrderItems {
		sum += v.ProductPrice * float32(v.Count)
	}

	return sum
}

func makeStorageItemsMap(items []*models.StorageItem) map[uint]*models.StorageItem {
	m := make(map[uint]*models.StorageItem)

	for _, v := range items {
		m[v.ProductID] = v
	}

	return m
}

func makeTransItemsMap(items []*models.StorageTransactionItem) map[uint]*models.StorageTransactionItem {
	m := make(map[uint]*models.StorageTransactionItem)

	for _, v := range items {
		m[v.ProductID] = v
	}

	return m
}
