package logic

import (
	"context"
	in "wallet_service/internal/app/interfaces"
	"wallet_service/internal/app/models"
)

func (s *PaymentService) purchaseProcessor(ctx context.Context, trans *in.Transaction) {
	code, err := s.processPurchase(ctx, trans)
	if err != nil {
		s.logger.Error("got process purchase error: ", err, code)

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
}

func (s *PaymentService) cancelationProcessor(ctx context.Context, trans *in.Transaction) {
	err := s.processCancelation(ctx, trans)
	if err != nil {
		s.logger.Error("got process cancellation error: ", err)
	}
}

func (s *PaymentService) invalidTransProcessor(trans *in.Transaction) {
	s.logger.Error("Invalid transaction type")

	trans.Type = models.Cancelation
	s.transactionsPipe <- trans
}
