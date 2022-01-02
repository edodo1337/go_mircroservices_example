package logic

import (
	"context"
	in "registry_service/internal/app/interfaces"
	"registry_service/internal/app/models"
	"time"
)

// Get orders list by user_id
func (s *OrdersService) GetOrdersList(ctx context.Context, userID uint) ([]*models.Order, error) {
	orders, err := s.ordersDAO.GetOrdersListByUserID(ctx, userID)

	return orders, err
}

// Entry point for making creating an order.
// Enriches new order data with pricing and redirects flow to NewOrdersPipeline
func (s *OrdersService) MakeOrder(
	ctx context.Context,
	makeOrderData in.MakeOrderDTO,
) error {
	productIDs := make([]uint, 0, 5)

	for _, item := range makeOrderData.OrderItems {
		productIDs = append(productIDs, item.ProductID)
	}

	productsPricesMap, err := s.productPricesDAO.GetProductPricesMap(ctx, productIDs)
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
	}

	return nil
}

// New orders pipeline processor.
// Gathers incoming item from ordersPipe channel
// and processes new order.
func (s *OrdersService) NewOrdersPipeline(
	ctx context.Context,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case orderData := <-s.newOrdersPipe:
			s.logger.Debug("OrderDATA:", *orderData)

			err := s.processNewOrder(ctx, orderData)
			if err != nil {
				s.logger.Error("Err process order:", err)
			}
		}
	}
}

// Handles new order, deals with data layer,
// sends new order msg to queue.
func (s *OrdersService) processNewOrder(
	ctx context.Context,
	newOrderData *in.NewOrderDTO,
) error {
	orderData := &in.CreateOrderDTO{UserID: newOrderData.UserID}

	order, err := s.ordersDAO.CreateOrder(ctx, orderData)
	if err != nil {
		return err
	}

	orderItemsData := make([]*in.CreateOrderItemDTO, 0, 5)
	for _, v := range newOrderData.OrderItems {
		orderItemsData = append(orderItemsData, &in.CreateOrderItemDTO{
			ProductID:    v.ProductID,
			Count:        v.Count,
			ProductPrice: v.ProductPrice,
		})
	}

	orderItems, err := s.orderItemsDAO.CreateOrderItemsBulk(ctx, order.ID, orderItemsData)
	if err != nil {
		// TODO dont delete but mark as rejected
		// or implement transactions
		s.ordersDAO.DeleteOrder(ctx, order.ID)

		return err
	}

	order.OrderItems = orderItems

	err = s.sendNewOrderMsg(ctx, order)
	if err != nil {
		// TODO dont delete but mark as rejected
		// or implement transactions
		s.ordersDAO.DeleteOrder(ctx, order.ID)

		return err
	}

	return nil
}

// Sends msg about new order to queue.
func (s *OrdersService) sendNewOrderMsg(ctx context.Context, order *models.Order) error {
	items := make([]in.NewOrderMsgItem, 0, 10)

	for _, v := range order.OrderItems {
		items = append(items, in.NewOrderMsgItem{
			OrderID:      order.ID,
			ProductID:    v.ProductID,
			Count:        v.Count,
			ProductPrice: v.ProductPrice,
		})
	}

	msg := &in.NewOrderMsg{
		UserID:     order.UserID,
		OrderID:    order.ID,
		OrderItems: items,
	}

	err := s.brokerClient.SendNewOrderMsg(ctx, msg)
	if err != nil {
		return err
	}

	return nil
}

func (s *OrdersService) ConsumeRejectedOrderMsgLoop(ctx context.Context) {
	ticker := time.NewTicker(s.consumeLoopTick)

	for {
		select {
		case <-ticker.C:
			msg, err := s.brokerClient.GetOrderRejectedMsg(ctx)
			if err == nil {
				s.logger.Debug("Kafka rejected order msg:", msg)

				_, updateErr := s.ordersDAO.UpdateOrderStatus(ctx, msg.OrderID, models.Rejected, msg.ReasonCode)
				if updateErr != nil {
					s.logger.Error("rejected order update err", updateErr)
				}
			} else {
				s.logger.Error("get order rejected msg err", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
