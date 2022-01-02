package logic

import (
	"context"
	in "storage_service/internal/app/interfaces"
	"storage_service/internal/app/models"
	"time"
)

// Entry point to reserve product items in storage
func (s *StorageService) MakeReservation(
	ctx context.Context,
	orderData *in.OrderDTO,
) error {
	items := make([]*in.TransactionItem, 0, 10)
	for _, v := range orderData.OrderItems {
		items = append(items, &in.TransactionItem{
			ProductID: v.ProductID,
			Count:     v.Count,
		})
	}

	trans := &in.Transaction{
		OrderID: orderData.OrderID,
		Items:   items,
		Type:    models.Reservation,
	}

	select {
	case s.transactionsPipe <- trans:
	case <-time.After(s.sendMsgTimeout * time.Second):
		return in.ErrNewTransactionTimeoutError
	}

	return nil
}

// Entry point to cancel reservation
func (s *StorageService) MakeCancelation(
	ctx context.Context,
	orderData in.CancelOrderDTO,
) error {
	transItems, err := s.storageTransactionsDAO.GetItemsByOrderID(ctx, orderData.OrderID)
	if err != nil {
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

	return nil
}

func (s *StorageService) ProcessTransactionsPipe(ctx context.Context) {
	for {
		select {
		case trans := <-s.transactionsPipe:
			switch trans.Type {
			case models.Reservation:
				err := s.processReservation(ctx, trans)
				if err != nil {
					s.logger.Error("got process reservation error")

					trans.Type = models.Cancelation
					s.transactionsPipe <- trans
				}
			case models.Cancelation:
				err := s.processCancelation(ctx, trans)
				if err != nil {
					s.logger.Error("got process cancellation error")
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

func (s *StorageService) processReservation(ctx context.Context, data *in.Transaction) error {
	existingTrans, err := s.storageTransactionsDAO.GetByOrderID(ctx, data.OrderID)
	if err != nil {
		return err
	}

	if existingTrans != nil {
		return nil
	}

	items := make([]*in.CreateStorageTransactionItemDTO, 0, 10)
	for _, v := range data.Items {
		items = append(items, &in.CreateStorageTransactionItemDTO{
			ProductID: v.ProductID,
			Count:     v.Count,
		})
	}

	newTrans, err := s.storageTransactionsDAO.Create(ctx, &in.CreateStorageTransactionDTO{
		OrderID: data.OrderID,
		Items:   items,
		Type:    data.Type,
	})
	if err != nil {
		return err
	}

	productIDs := make([]uint, 0, 10)
	for _, v := range data.Items {
		productIDs = append(productIDs, v.ProductID)
	}

	storageItems, err := s.storageItemsDAO.GetListByProductIDs(ctx, productIDs)
	if err != nil {
		return err
	}

	itemsMap := makeStorageItemsMap(storageItems)
	transMap := makeTransItemsMap(newTrans.Items)

	for k, v := range transMap {
		product, exists := itemsMap[k]
		if !exists {
			return in.ErrProductNotFoundByID
		}

		if product.Count < v.Count {
			return in.ErrOutOfStock
		}

		product.Count -= v.Count
	}

	if err := s.storageItemsDAO.UpdateCountBulk(ctx, storageItems); err != nil {
		return err
	}

	return nil
}

func (s *StorageService) processCancelation(ctx context.Context, data *in.Transaction) error {
	existingTrans, err := s.storageTransactionsDAO.GetByOrderID(ctx, data.OrderID)
	if err != nil {
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

	newTrans, err := s.storageTransactionsDAO.Create(ctx, &in.CreateStorageTransactionDTO{
		OrderID: data.OrderID,
		Items:   items,
		Type:    data.Type,
	})
	if err != nil {
		return err
	}

	productIDs := make([]uint, 0, 10)
	for _, v := range data.Items {
		productIDs = append(productIDs, v.ProductID)
	}

	storageItems, err := s.storageItemsDAO.GetListByProductIDs(ctx, productIDs)
	if err != nil {
		return err
	}

	itemsMap := makeStorageItemsMap(storageItems)
	transMap := makeTransItemsMap(newTrans.Items)

	for k, v := range transMap {
		product, exists := itemsMap[k]
		if !exists {
			return in.ErrProductNotFoundByID
		}

		if product.Count < v.Count {
			return in.ErrOutOfStock
		}

		product.Count += v.Count
	}

	if err := s.storageItemsDAO.UpdateCountBulk(ctx, storageItems); err != nil {
		return err
	}

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
			if err == nil {
				s.logger.Debug("Kafka rejected order msg:", msg)

				cancelOrderData := &in.CancelOrderDTO{
					OrderID: msg.OrderID,
					UserID:  msg.UserID,
					Cost:    msg.Cost,
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
