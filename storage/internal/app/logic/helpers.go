package logic

import (
	"context"
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

func makeTransItemsMap(items []*in.CreateStorageTransactionItemDTO) map[uint]*in.CreateStorageTransactionItemDTO {
	m := make(map[uint]*in.CreateStorageTransactionItemDTO)

	for _, v := range items {
		m[v.ProductID] = v
	}

	return m
}

func (s *StorageService) sendSuccessMsg(ctx context.Context, data *in.Transaction) error {
	err := s.brokerClient.SendReservationSuccess(ctx, &in.OrderSuccessMsg{
		OrderID: data.OrderID,
		Service: in.Storage,
	})

	return err
}

func (s *StorageService) sendRejectedMsg(ctx context.Context, reasonCode models.CancelationReason, data *in.Transaction) error {
	s.logger.Info("Kafka Send rejected message: ", data, reasonCode)

	err := s.brokerClient.SendOrderRejectedMsg(ctx, &in.OrderRejectedMsg{
		OrderID:    data.OrderID,
		UserID:     data.UserID,
		ReasonCode: reasonCode,
		Service:    in.Storage,
	})

	return err
}
