package logic

import (
	"context"
	in "wallet_service/internal/app/interfaces"
	"wallet_service/internal/app/models"
)

func calcOrderSum(orderData *in.OrderDTO) float32 {
	var sum float32
	for _, v := range orderData.OrderItems {
		sum += v.ProductPrice * float32(v.Count)
	}

	return sum
}

func (s *PaymentService) sendSuccessMsg(ctx context.Context, data *in.Transaction) error {
	err := s.brokerClient.SendPurchaseSuccess(ctx, &in.OrderSuccessMsg{
		OrderID: data.OrderID,
		Service: in.Storage,
	})

	return err
}

func (s *PaymentService) sendRejectedMsg(ctx context.Context, reasonCode models.CancelationReason, data *in.Transaction) error {
	err := s.brokerClient.SendOrderRejectedMsg(ctx, &in.OrderRejectedMsg{
		OrderID:    data.OrderID,
		ReasonCode: reasonCode,
		Service:    in.Storage,
	})

	return err
}
