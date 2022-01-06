package logic

import (
	"context"
	in "registry_service/internal/app/interfaces"
	"registry_service/internal/app/models"
	"sync"
	"time"
)

// Get products list
func (s *OrdersService) GetProductList(ctx context.Context) ([]*models.Product, error) {
	products, err := s.productPricesDAO.GetList(ctx)

	return products, err
}

// Get orders list by user_id
func (s *OrdersService) GetOrdersList(ctx context.Context, userID uint) ([]*models.Order, error) {
	orders, err := s.ordersDAO.GetListByUserID(ctx, userID)

	return orders, err
}

// Entry point for making creating an order.
// Enriches new order data with pricing and redirects flow to NewOrdersPipeline
func (s *OrdersService) MakeOrder(
	ctx context.Context,
	makeOrderData *in.MakeOrderDTO,
) error {
	s.logger.Info("Making order")

	productIDs := make([]uint, 0, 5)

	for _, item := range makeOrderData.OrderItems {
		productIDs = append(productIDs, item.ProductID)
	}

	productsPricesMap, err := s.productPricesDAO.GetMap(ctx, productIDs)
	if err != nil {
		return err
	}

	orderItemsDTOs := enrichOrderItemsDataWithPrices(productsPricesMap, makeOrderData.OrderItems)

	newOrderDTO := &in.NewOrderDTO{
		UserID:     makeOrderData.UserID,
		OrderItems: orderItemsDTOs,
	}

	select {
	case s.newOrdersPipe <- newOrderDTO:
	case <-time.After(s.sendMsgTimeout * time.Second):
		return in.ErrNewOrderTimeout
	case <-ctx.Done():
		return nil
	}

	s.logger.Info("Making order success")

	return nil
}

// Entry point for orders cancelation.
func (s *OrdersService) MakeCancelation(
	ctx context.Context,
	cancelData *in.OrderRejectedMsg,
) error {
	s.logger.Info("Making cancelation")

	select {
	case s.rejectedOrdersPipe <- cancelData:
	case <-time.After(s.sendMsgTimeout * time.Second):
		return in.ErrNewOrderTimeout
	case <-ctx.Done():
		return nil
	}

	s.logger.Info("Making cancelation: send msg to chan success")

	return nil
}

// Entry point for mark order success step:
// Paid, Reserved ...
func (s *OrdersService) MarkSuccessStep(
	ctx context.Context,
	successData *in.OrderSuccessMsg,
) error {
	s.logger.Info("Marking order as successful")

	select {
	case s.successOrdersPipe <- successData:
	case <-time.After(s.sendMsgTimeout * time.Second):
		return in.ErrNewOrderTimeout
	case <-ctx.Done():
		return nil
	}

	s.logger.Info("Making cancelation: send msg to chan success")

	return nil
}

// New events pipeline processor.
// Gathers incoming item from pipe channels
// and processes new order, cancelation and success msg.
func (s *OrdersService) EventPipeProcessor(
	ctx context.Context,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		select {
		case orderData, ok := <-s.newOrdersPipe:
			if !ok {
				return
			}

			s.logger.Info("New order data: ", *orderData)

			err := s.processNewOrder(ctx, orderData)
			if err != nil {
				s.logger.Error("Err process order: ", err)
			}
		case cancelData, ok := <-s.rejectedOrdersPipe:
			if !ok {
				return
			}

			s.logger.Info("New cancelation data: ", *cancelData)

			err := s.processCancelation(ctx, cancelData)
			if err != nil {
				s.logger.Error("Err process cancelation: ", err)
			}
		case successData, ok := <-s.successOrdersPipe:
			if !ok {
				return
			}

			s.logger.Info("New success order data: ", *successData)

			err := s.processSuccess(ctx, successData)
			if err != nil {
				s.logger.Error("Err process success order: ", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *OrdersService) ConsumeRejectedOrderMsgLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(s.consumeLoopTick)

	for {
		select {
		case <-ticker.C:
			msg, err := s.brokerClient.GetOrderRejectedMsg(ctx)
			if err != nil {
				s.logger.Error("got order rejected msg err: ", err)
			}

			if msg == nil {
				continue
			}

			if msg.Service == in.Registry {
				s.logger.Info("Got message for registry. Skip")

				continue
			}

			s.logger.Info("Kafka rejected order msg: ", msg)

			errCancel := s.MakeCancelation(ctx, msg)
			if errCancel != nil {
				s.logger.Errorf("Order cancelation error: %v", errCancel)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (s *OrdersService) ConsumeSuccessMsgLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(s.consumeLoopTick)

	for {
		select {
		case <-ticker.C:
			msg, err := s.brokerClient.GetSuccessMsg(ctx)
			if err != nil {
				s.logger.Error("Error get success order msg from kafka: ", err)
			}

			if msg == nil {
				continue
			}

			s.logger.Info("New success order msg", msg)

			errCancel := s.MarkSuccessStep(ctx, msg)
			if errCancel != nil {
				s.logger.Errorf("Make success steo error: %v", errCancel)
			}
		case <-ctx.Done():
			return
		}
	}
}
