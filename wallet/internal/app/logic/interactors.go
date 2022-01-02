package logic

import (
	"context"
	"time"
	in "wallet_service/internal/app/interfaces"
	"wallet_service/internal/app/models"
)

// Entry point to make purchase
func (s *PaymentService) MakePurchase(
	ctx context.Context,
	orderData *in.OrderDTO,
) error {
	wallet, err := s.walletsDAO.GetByUserID(ctx, orderData.UserID)
	if err != nil {
		return err
	}

	cost := calcOrderSum(orderData)

	trans := &in.Transaction{
		Cost:    cost,
		OrderID: orderData.OrderID,
		Wallet:  wallet,
		Type:    models.Purchase,
	}

	select {
	case s.transactionsPipe <- trans:
	case <-time.After(s.sendMsgTimeout * time.Second):
		return in.ErrNewTransactionTimeoutError
	}

	return nil
}

// Entry point to cancel purchase
func (s *PaymentService) MakeCancelation(
	ctx context.Context,
	orderData in.CancelOrderDTO,
) error {
	wallet, err := s.walletsDAO.GetByUserID(ctx, orderData.UserID)
	if err != nil {
		return err
	}

	trans := &in.Transaction{
		Cost:    orderData.Cost,
		OrderID: orderData.OrderID,
		Wallet:  wallet,
		Type:    models.Cancelation,
	}

	select {
	case s.transactionsPipe <- trans:
	case <-time.After(s.sendMsgTimeout * time.Second):
		return in.ErrNewTransactionTimeoutError
	}

	return nil
}

func (s *PaymentService) ProcessTransactionsPipe(ctx context.Context) {
	for {
		select {
		case trans := <-s.transactionsPipe:
			switch trans.Type {
			case models.Purchase:
				err := s.processPurchase(ctx, trans)
				if err != nil {
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

func (s *PaymentService) processPurchase(ctx context.Context, trans *in.Transaction) error {
	existingTrans, err := s.walletsTransactionsDAO.GetByOrderID(ctx, trans.OrderID)
	if err != nil {
		return err
	}

	if existingTrans != nil {
		return nil
	}

	newTrans, err := s.walletsTransactionsDAO.Create(ctx, &in.CreateWalletTransactionDTO{
		WalletID: trans.Wallet.ID,
		OrderID:  trans.OrderID,
		Cost:     trans.Cost,
		Type:     trans.Type,
	})
	if err != nil {
		return err
	}

	trans.Wallet.Balance -= newTrans.Cost
	if _, err := s.walletsDAO.UpdateBalance(ctx, trans.Wallet); err != nil {
		return err
	}

	return nil
}

func (s *PaymentService) processCancelation(ctx context.Context, trans *in.Transaction) error {
	existingTrans, err := s.walletsTransactionsDAO.GetByOrderID(ctx, trans.OrderID)
	if err != nil {
		return err
	}

	if existingTrans != nil && existingTrans.Type == models.Cancelation {
		return nil
	}

	newTrans, err := s.walletsTransactionsDAO.Create(ctx, &in.CreateWalletTransactionDTO{
		WalletID: trans.Wallet.ID,
		OrderID:  trans.OrderID,
		Cost:     trans.Cost,
		Type:     trans.Type,
	})
	if err != nil {
		return err
	}

	trans.Wallet.Balance += newTrans.Cost
	if _, err := s.walletsDAO.UpdateBalance(ctx, trans.Wallet); err != nil {
		return err
	}

	return nil
}

func (s *PaymentService) ConsumeNewOrderMsgLoop(ctx context.Context) {
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
						Count:        v.Count,
						ProductPrice: v.ProductPrice,
					})
				}

				orderData := in.OrderDTO{
					OrderID:    msg.OrderID,
					UserID:     msg.UserID,
					OrderItems: orderItemsData,
				}

				if purchaseErr := s.MakePurchase(ctx, &orderData); purchaseErr != nil {
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

func (s *PaymentService) ConsumeRejectedOrderMsgLoop(ctx context.Context) {
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
						Count:        v.Count,
						ProductPrice: v.ProductPrice,
					})
				}

				orderData := in.OrderDTO{
					OrderID:    msg.OrderID,
					UserID:     msg.UserID,
					OrderItems: orderItemsData,
				}

				if purchaseErr := s.MakePurchase(ctx, &orderData); purchaseErr != nil {
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
