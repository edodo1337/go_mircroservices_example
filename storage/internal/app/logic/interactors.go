package logic

import (
	"context"
	"errors"
	in "storage_service/internal/app/interfaces"
	"storage_service/internal/app/models"
	"time"
)

// Entry point to reserve product items in storage
func (s *StorageService) MakeReservation(
	ctx context.Context,
	orderData *in.OrderDTO,
) error {
	s.logger.Info("Making reservation", orderData)

	items := make([]*in.TransactionItem, 0, 10)
	for _, v := range orderData.OrderItems {
		items = append(items, &in.TransactionItem{
			ProductID: v.ProductID,
			Count:     v.Count,
		})
	}

	trans := &in.Transaction{
		OrderID: orderData.OrderID,
		UserID:  orderData.UserID,
		Items:   items,
		Type:    models.Reservation,
	}

	select {
	case s.transactionsPipe <- trans:
	case <-time.After(s.sendMsgTimeout * time.Second):
		return in.ErrNewTransactionTimeoutError
	}

	s.logger.Info("Reservation success")

	return nil
}

// Entry point to cancel reservation
func (s *StorageService) MakeCancelation(
	ctx context.Context,
	orderData in.CancelOrderDTO,
) error {
	s.logger.Info("Making cancelation")

	transItems, err := s.storageTransactionsDAO.GetItemsByOrderID(ctx, orderData.OrderID)
	if err != nil && !errors.Is(err, in.ErrTransNotFound) {
		return err
	}

	items := make([]*in.TransactionItem, 0, 10)
	for _, v := range transItems {
		items = append(items, &in.TransactionItem{
			ProductID: v.ProductID,
			Count:     v.Count,
		})
	}

	trans := &in.Transaction{
		OrderID: orderData.OrderID,
		Items:   items,
		Type:    models.Cancelation,
	}

	select {
	case s.transactionsPipe <- trans:
	case <-time.After(s.sendMsgTimeout * time.Second):
		return in.ErrNewTransactionTimeoutError
	}

	s.logger.Info("Cancelation success")

	return nil
}

func (s *StorageService) ProcessTransactionsPipe(ctx context.Context) {
	for {
		select {
		case trans := <-s.transactionsPipe:
			switch trans.Type {
			case models.Reservation:
				code, err := s.processReservation(ctx, trans)
				if err != nil {
					s.logger.Error("got process reservation error: ", err, code)

					trans.Type = models.Cancelation
					s.transactionsPipe <- trans

					errSend := s.sendRejectedMsg(ctx, code, trans)
					if errSend != nil {
						s.logger.Error("send rejected msg error: ", errSend)
					}
				}

				errSend := s.sendSuccessMsg(ctx, trans)
				if errSend != nil {
					s.logger.Error("send success msg error: ", errSend)
				}
			case models.Cancelation:
				err := s.processCancelation(ctx, trans)
				if err != nil {
					s.logger.Error("got process cancellation error: ", err)
				}
			default:
				s.logger.Error("Invalid transaction type")

				trans.Type = models.Cancelation
				s.transactionsPipe <- trans
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *StorageService) processReservation(ctx context.Context, data *in.Transaction) (models.CancelationReason, error) {
	s.logger.Info("Processing reservation")

	existingTrans, err := s.storageTransactionsDAO.GetByOrderID(ctx, data.OrderID)
	if err != nil && !errors.Is(err, in.ErrTransNotFound) {
		s.logger.Error("got process reservation err: ", err)

		return models.InternalError, err
	}

	if existingTrans != nil {
		return models.OK, nil
	}

	items := make([]*in.CreateStorageTransactionItemDTO, 0, 10)
	for _, v := range data.Items {
		items = append(items, &in.CreateStorageTransactionItemDTO{
			ProductID: v.ProductID,
			Count:     v.Count,
		})
	}

	productIDs := make([]uint, 0, 10)
	for _, v := range data.Items {
		productIDs = append(productIDs, v.ProductID)
	}

	storageItems, err := s.storageItemsDAO.GetListByProductIDs(ctx, productIDs)
	if err != nil {
		return models.InternalError, err
	}

	storageItemsMap := makeStorageItemsMap(storageItems)
	transItemsMap := makeTransItemsMap(items)

	for k, v := range transItemsMap {
		product, exists := storageItemsMap[k]
		if !exists {
			return models.InternalError, in.ErrProductNotFoundByID
		}

		if product.Count < v.Count {
			return models.OutOfStock, in.ErrOutOfStock
		}

		product.Count -= v.Count
	}

	_, err = s.storageTransactionsDAO.Create(ctx, &in.CreateStorageTransactionDTO{
		OrderID: data.OrderID,
		Items:   items,
		Type:    data.Type,
	})
	if err != nil {
		return models.InternalError, err
	}

	if errUpd := s.storageItemsDAO.UpdateCountBulk(ctx, storageItems); errUpd != nil {
		s.logger.Error("got process reservation update count err:", errUpd)

		return models.InternalError, errUpd
	}

	s.logger.Info("Processing reservation success")

	err = s.brokerClient.SendReservationSuccess(ctx, &in.OrderSuccessMsg{
		OrderID: data.OrderID,
		Service: in.Storage,
	})
	if err != nil {
		s.logger.Info("Send reservation msg error")
	}

	s.logger.Info("Processing reservation success")

	return models.OK, nil
}

func (s *StorageService) processCancelation(ctx context.Context, data *in.Transaction) error {
	s.logger.Info("Processing cancelation")

	existingTrans, err := s.storageTransactionsDAO.GetByOrderID(ctx, data.OrderID)

	if errors.Is(err, in.ErrTransNotFound) {
		s.logger.Info("Processing cancelation success: order not found")

		return nil
	} else if err != nil {
		return err
	}

	if existingTrans != nil && existingTrans.Type == models.Cancelation {
		return nil
	}

	items := make([]*in.CreateStorageTransactionItemDTO, 0, 10)
	for _, v := range data.Items {
		items = append(items, &in.CreateStorageTransactionItemDTO{
			ProductID: v.ProductID,
			Count:     v.Count,
		})
	}

	productIDs := make([]uint, 0, 10)
	for _, v := range data.Items {
		productIDs = append(productIDs, v.ProductID)
	}

	storageItems, err := s.storageItemsDAO.GetListByProductIDs(ctx, productIDs)
	if err != nil {
		return err
	}

	storageItemsMap := makeStorageItemsMap(storageItems)
	transItemsMap := makeTransItemsMap(items)

	for k, v := range transItemsMap {
		product, exists := storageItemsMap[k]
		if !exists {
			return in.ErrProductNotFoundByID
		}

		product.Count += v.Count
	}

	_, err = s.storageTransactionsDAO.Create(ctx, &in.CreateStorageTransactionDTO{
		OrderID: data.OrderID,
		Items:   items,
		Type:    data.Type,
	})
	if err != nil {
		return err
	}

	if err := s.storageItemsDAO.UpdateCountBulk(ctx, storageItems); err != nil {
		s.logger.Error("got process cancelation update count err:", err)

		return err
	}

	s.logger.Info("Processing cancelation success")

	return nil
}

func (s *StorageService) ConsumeNewOrderMsgLoop(ctx context.Context) {
	ticker := time.NewTicker(s.consumeLoopTick)

	for {
		select {
		case <-ticker.C:
			msg, err := s.brokerClient.GetNewOrderMsg(ctx)
			if err == nil {
				s.logger.Debug("Kafka new order msg:", msg)

				orderItemsData := make([]*in.OrderItemDTO, 0, 10)
				for _, v := range msg.OrderItems {
					orderItemsData = append(orderItemsData, &in.OrderItemDTO{
						ProductID:    v.ProductID,
						Count:        uint16(v.Count),
						ProductPrice: v.ProductPrice,
					})
				}

				orderData := in.OrderDTO{
					OrderID:    msg.OrderID,
					UserID:     msg.UserID,
					OrderItems: orderItemsData,
				}

				if purchaseErr := s.MakeReservation(ctx, &orderData); purchaseErr != nil {
					s.logger.Error("new order make purchase err", purchaseErr)
				}
			} else {
				s.logger.Error("got new order msg err", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *StorageService) ConsumeRejectedOrderMsgLoop(ctx context.Context) {
	ticker := time.NewTicker(s.consumeLoopTick)

	for {
		select {
		case <-ticker.C:
			msg, err := s.brokerClient.GetOrderRejectedMsg(ctx)

			if msg.Service == in.Storage {
				s.logger.Info("Got message for storage. Skip")

				continue
			}

			if err == nil {
				s.logger.Debug("Kafka rejected order msg:", msg)

				cancelOrderData := &in.CancelOrderDTO{
					OrderID: msg.OrderID,
					UserID:  msg.UserID,
				}

				if cancelErr := s.MakeCancelation(ctx, *cancelOrderData); cancelErr != nil {
					s.logger.Error("rejected order update err", cancelErr)
				}
			} else {
				s.logger.Error("get order rejected msg err", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
