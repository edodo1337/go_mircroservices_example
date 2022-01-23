package logic

import (
	"context"
	in "registry_service/internal/app/interfaces"
)

func (s *OrdersService) newOrderProcessor(ctx context.Context, orderData *in.NewOrderDTO) {
	s.logger.Info("New order data: ", *orderData)

	err := s.processNewOrder(ctx, orderData)
	if err != nil {
		s.logger.Error("Err process order: ", err)
	}
}

func (s *OrdersService) rejectedOrderProcessor(ctx context.Context, cancelData *in.OrderRejectedMsg) {
	s.logger.Info("New cancelation data: ", *cancelData)

	err := s.processCancelation(ctx, cancelData)
	if err != nil {
		s.logger.Error("Err process cancelation: ", err)
	}
}

func (s *OrdersService) successOrderProcessor(ctx context.Context, successData *in.OrderSuccessMsg) {
	s.logger.Info("New success order data: ", *successData)

	err := s.processSuccess(ctx, successData)
	if err != nil {
		s.logger.Error("Err process success order: ", err)
	}
}
