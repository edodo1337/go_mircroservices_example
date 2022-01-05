package logic

import (
	"context"
	in "registry_service/internal/app/interfaces"
	"registry_service/internal/app/models"
)

// Helper func for products prices enrichment.
func enrichOrderItemsDataWithPrices(
	productPricesMap in.ProductPricesMap,
	orderItemsData []*in.MakeOrderItemDTO,
) []*in.NewOrderItemDTO {
	orderItemsDTOs := make([]*in.NewOrderItemDTO, 0, 10)

	for _, item := range orderItemsData {
		val := productPricesMap[item.ProductID]

		orderItemsDTOs = append(orderItemsDTOs, &in.NewOrderItemDTO{
			ProductID:    item.ProductID,
			Count:        item.Count,
			ProductPrice: val,
		})
	}

	return orderItemsDTOs
}

// Processes success order msgs
func (s *OrdersService) processSuccess(ctx context.Context, msg *in.OrderSuccessMsg) error {
	s.logger.Infof("Processing success orders: %v", msg)

	order, err := s.ordersDAO.GetByID(ctx, msg.OrderID)
	if err != nil {
		s.logger.Errorf("Processing success orders: get by id err %v", err)

		return err
	}

	switch msg.Service {
	case in.Storage:
		if order.Status == models.Paid {
			order.Status = models.Completed
		} else {
			order.Status = models.Reserved
		}
	case in.Wallet:
		if order.Status == models.Reserved {
			order.Status = models.Completed
		} else {
			order.Status = models.Paid
		}
	case in.Registry:
	default:
	}

	_, err = s.ordersDAO.UpdateStatus(ctx, order.ID, order.Status, models.OK)
	if err != nil {
		s.logger.Errorf("Processing success orders: update status err: %v", err)

		return err
	}

	s.logger.Info("Processing success orders: ok")

	return nil
}

// Processes cancelation msgs
func (s *OrdersService) processCancelation(ctx context.Context, msg *in.OrderRejectedMsg) error {
	s.logger.Infof("Processing rejected: %v", msg)

	_, updateErr := s.ordersDAO.UpdateStatus(ctx, msg.OrderID, models.Rejected, msg.ReasonCode)
	if updateErr != nil {
		s.logger.Errorf("Processing rejected: order update err %v", updateErr)

		return updateErr
	}

	s.logger.Info("Processing rejected success")

	return nil
}

// Sends msg about new order to queue.
func (s *OrdersService) sendNewOrderMsg(ctx context.Context, order *models.Order) error {
	s.logger.Logger.Debug("Send new order msg")

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

// Handles new order, deals with data layer,
// sends new order msg to queue.
func (s *OrdersService) processNewOrder(
	ctx context.Context,
	newOrderData *in.NewOrderDTO,
) error {
	orderData := &in.CreateOrderDTO{UserID: newOrderData.UserID}

	order, err := s.ordersDAO.Create(ctx, orderData)
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

	orderItems, err := s.orderItemsDAO.CreateBulk(ctx, order.ID, orderItemsData)
	if err != nil {
		_, errUpd := s.ordersDAO.UpdateStatus(
			ctx,
			order.ID,
			models.Rejected,
			models.InternalError,
		)
		if errUpd != nil {
			return errUpd
		}

		return err
	}

	order.OrderItems = orderItems

	err = s.sendNewOrderMsg(ctx, order)
	if err != nil {
		_, errUpd := s.ordersDAO.UpdateStatus(
			ctx,
			order.ID,
			models.Rejected,
			models.InternalError,
		)

		if errUpd != nil {
			return errUpd
		}

		return err
	}

	return nil
}
