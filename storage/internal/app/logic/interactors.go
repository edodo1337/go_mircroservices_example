package logic

import (
	"context"
	"errors"
	in "storage_service/internal/app/interfaces"
	"storage_service/internal/app/models"
	"sync"
	"time"
)

// Entry point to reserve product items in storage
func (s *StorageService) MakeReservation(
	ctx context.Context,
	orderData *in.OrderDTO,
) error {
	s.logger.Infof("Making reservation: %v", orderData)

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

	s.logger.Info("Reservation: send msg to chan success")

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

	s.logger.Info("Cancelation: send msg to chan success")

	return nil
}

func (s *StorageService) EventPipeProcessor(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case trans, ok := <-s.transactionsPipe:
			if !ok {
				return
			}

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

func (s *StorageService) ConsumeNewOrderMsgLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(s.consumeLoopTick)

	for {
		select {
		case <-ticker.C:
			msg, err := s.brokerClient.GetNewOrderMsg(ctx)
			if err != nil {
				s.logger.Error("got new order msg err: ", err)
			}

			if msg == nil {
				continue
			}

			s.logger.Debug("Kafka new order msg: ", msg)

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
				s.logger.Error("new order make purchase err: ", purchaseErr)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *StorageService) ConsumeRejectedOrderMsgLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(s.consumeLoopTick)

	for {
		select {
		case <-ticker.C:
			msg, err := s.brokerClient.GetOrderRejectedMsg(ctx)
			if err != nil {
				s.logger.Error("get order rejected msg err: ", err)
			}

			if msg == nil {
				continue
			}

			s.logger.Debug("Kafka rejected order msg: ", msg)

			if msg.Service == in.Storage {
				s.logger.Info("Got message for storage. Skip")

				continue
			}

			cancelOrderData := &in.CancelOrderDTO{
				OrderID: msg.OrderID,
				UserID:  msg.UserID,
			}

			if cancelErr := s.MakeCancelation(ctx, *cancelOrderData); cancelErr != nil {
				s.logger.Error("rejected order update err: ", cancelErr)
			}
		case <-ctx.Done():
			return
		}
	}
}
