package logic

import (
	"context"
	"errors"
	"sync"
	"time"
	in "wallet_service/internal/app/interfaces"
	"wallet_service/internal/app/models"
)

// Entry point to make purchase
func (s *PaymentService) MakePurchase(
	ctx context.Context,
	orderData *in.OrderDTO,
) error {
	s.logger.Info("Making purchase")

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

	s.logger.Info("Purchase: send msg to chan success")

	return nil
}

// Entry point to cancel purchase
func (s *PaymentService) MakeCancelation(
	ctx context.Context,
	orderData in.CancelOrderDTO,
) error {
	s.logger.Info("Making cancelation: ", orderData)

	wallet, err := s.walletsDAO.GetByUserID(ctx, orderData.UserID)
	if err != nil {
		s.logger.Error("Cancelation wallet get by user id err: ", err)

		return err
	}

	oldTrans, err := s.walletsTransactionsDAO.GetByOrderID(ctx, orderData.OrderID)
	if err != nil {
		s.logger.Error("Cancelation wallet old trans not found err: ", err)

		return err
	}

	trans := &in.Transaction{
		Cost:    oldTrans.Cost,
		OrderID: orderData.OrderID,
		Wallet:  wallet,
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

func (s *PaymentService) EventPipeProcessor(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case trans, ok := <-s.transactionsPipe:
			if !ok {
				continue
			}

			switch trans.Type {
			case models.Purchase:
				go s.purchaseProcessor(ctx, trans)

			case models.Cancelation:
				go s.cancelationProcessor(ctx, trans)

			default:
				go s.invalidTransProcessor(trans)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *PaymentService) processPurchase(ctx context.Context, trans *in.Transaction) (models.CancelationReason, error) {
	s.logger.Info("Processing purchase")

	existingTrans, err := s.walletsTransactionsDAO.GetByOrderID(ctx, trans.OrderID)
	if err != nil && !errors.Is(err, in.ErrTransNotFound) {
		s.logger.Error("got process purchase err: ", err)

		return models.InternalError, err
	}

	if existingTrans != nil {
		return models.OK, nil
	}

	newTrans, err := s.walletsTransactionsDAO.Create(ctx, &in.CreateWalletTransactionDTO{
		WalletID: trans.Wallet.ID,
		OrderID:  trans.OrderID,
		Cost:     trans.Cost,
		Type:     trans.Type,
	})
	if err != nil {
		return models.InternalError, err
	}

	if trans.Wallet.Balance < newTrans.Cost {
		return models.NotEnoughMoney, in.ErrNotEnoughMoney
	}

	trans.Wallet.Balance -= newTrans.Cost
	if _, err := s.walletsDAO.UpdateBalance(ctx, trans.Wallet); err != nil {
		return models.InternalError, err
	}

	s.logger.Info("Processing purchase success")

	return models.OK, nil
}

func (s *PaymentService) processCancelation(ctx context.Context, trans *in.Transaction) error {
	s.logger.Info("Processing cancelation")

	existingTrans, err := s.walletsTransactionsDAO.GetByOrderID(ctx, trans.OrderID)
	if err != nil && !errors.Is(err, in.ErrTransNotFound) {
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

	s.logger.Info("Processing cancelation success")

	return nil
}

func (s *PaymentService) ConsumeNewOrderMsgLoop(ctx context.Context, wg *sync.WaitGroup) {
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
				s.logger.Error("new order make purchase err: ", purchaseErr)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *PaymentService) ConsumeRejectedOrderMsgLoop(ctx context.Context, wg *sync.WaitGroup) {
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

			s.logger.Debug("Kafka rejected order msg:", msg)

			if msg.Service == in.Wallet {
				s.logger.Info("Got message for wallet. Skip")

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
