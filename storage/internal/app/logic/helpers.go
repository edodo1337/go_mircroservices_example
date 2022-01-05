package logic

import (
	"context"
	"errors"
	in "storage_service/internal/app/interfaces"
	"storage_service/internal/app/models"
)

func makeStorageItemsMap(items []*models.StorageItem) map[uint]*models.StorageItem {
	m := make(map[uint]*models.StorageItem)

	for _, v := range items {
		m[v.ProductID] = v
	}

	return m
}

func makeTransItemsMap(items []*in.CreateStorageTransactionItemDTO) map[uint]*in.CreateStorageTransactionItemDTO {
	m := make(map[uint]*in.CreateStorageTransactionItemDTO)

	for _, v := range items {
		m[v.ProductID] = v
	}

	return m
}

func (s *StorageService) sendSuccessMsg(ctx context.Context, data *in.Transaction) error {
	err := s.brokerClient.SendReservationSuccess(ctx, &in.OrderSuccessMsg{
		OrderID: data.OrderID,
		Service: in.Storage,
	})

	return err
}

func (s *StorageService) sendRejectedMsg(ctx context.Context, reasonCode models.CancelationReason, data *in.Transaction) error {
	s.logger.Info("Kafka Send rejected message: ", data, reasonCode)

	err := s.brokerClient.SendOrderRejectedMsg(ctx, &in.OrderRejectedMsg{
		OrderID:    data.OrderID,
		UserID:     data.UserID,
		ReasonCode: reasonCode,
		Service:    in.Storage,
	})

	return err
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

	s.logger.Info("Processing reservation: update count success")

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
		s.logger.Error("got process cancelation update count err: ", err)

		return err
	}

	s.logger.Info("Processing cancelation success")

	return nil
}
